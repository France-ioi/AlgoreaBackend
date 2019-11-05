package database

import (
	"fmt"
)

// UserStore implements database operations on `users`
type UserStore struct {
	*DataStore
}

const deleteWithTrapsBatchSize = 1000

// DeleteTemporaryWithTraps deletes temporary users who don't have active sessions.
// It also removes linked rows in the tables:
// 1. [`users_threads`, `users_answers`, `users_items`, `filters`, `sessions`, `refresh_tokens`]
//    having `user_group_id` = `users.group_id`;
// 2. [`permissions_granted`, `permissions_generated`, `groups_attempts`, `groups_login_prefixes`]
//    having `group_id` = `users.group_id` or `group_id` = `users.owned_group_id`;
// 3. `groups_groups` having `parent_group_id` or `child_group_id` equal
//    to one of `users.group_id`/`users.owned_group_id`;
// 4. `groups_ancestors` having `ancestor_group_id` or `child_group_id` equal
//    to one of `users.group_id`/`users.owned_group_id`;
// 5. [`groups_propagate`, `groups`] having `id` equal to one of
//    `users.group_id`/`users.owned_group_id`.
func (s *UserStore) DeleteTemporaryWithTraps() (err error) {
	defer recoverPanics(&err)

	s.executeBatchesInTransactions(func(store *DataStore) int {
		userScope := store.Users().
			Joins("LEFT JOIN sessions ON sessions.user_group_id = users.group_id AND NOW() < sessions.expires_at").
			Where("sessions.user_group_id IS NULL").Where("temp_user = 1")
		return store.Users().deleteWithTraps(userScope)
	})
	return nil
}

// DeleteWithTraps deletes a given user. It also removes linked rows in the same way as DeleteTemporaryWithTraps.
func (s *UserStore) DeleteWithTraps(user *User) (err error) {
	return s.InTransaction(func(store *DataStore) error {
		deleteOneBatchOfUsers(store.DB, []int64{user.GroupID}, []*int64{user.OwnedGroupID})
		store.GroupGroups().createNewAncestors()
		return nil
	})
}

func (s *UserStore) executeBatchesInTransactions(f func(store *DataStore) int) {
	for {
		var usersDeleted int
		mustNotBeError(s.InTransaction(func(store *DataStore) error {
			usersDeleted = f(store)
			return nil
		}))
		if usersDeleted < deleteWithTrapsBatchSize {
			return
		}
	}
}

// deleteWithTraps deletes the first deleteWithTrapsBatchSize users satisfying the scope's conditions
// and all the users' stuff
func (s *UserStore) deleteWithTraps(userScope *DB) int {
	userScope.mustBeInTransaction()

	userGroupsIDs := make([]int64, 0, deleteWithTrapsBatchSize)
	ownedGroupsIDs := make([]*int64, 0, deleteWithTrapsBatchSize) // can be NULL

	mustNotBeError(
		userScope.WithWriteLock().Select("group_id, owned_group_id").Limit(deleteWithTrapsBatchSize).
			ScanIntoSlices(&userGroupsIDs, &ownedGroupsIDs).Error())

	if len(userGroupsIDs) == 0 {
		return 0
	}

	deleteOneBatchOfUsers(userScope, userGroupsIDs, ownedGroupsIDs)
	s.GroupGroups().createNewAncestors()

	return len(userGroupsIDs)
}

func deleteOneBatchOfUsers(db *DB, userGroupsIDs []int64, ownedGroupsIDs []*int64) {
	db.mustBeInTransaction()

	allGroups := make([]*int64, 0, len(userGroupsIDs)+len(ownedGroupsIDs))
	for _, id := range userGroupsIDs {
		id := id
		allGroups = append(allGroups, &id)
	}
	allGroups = append(allGroups, ownedGroupsIDs...)
	for _, table := range [...]string{
		"users_threads", "users_answers", "users_items", "filters", "sessions", "refresh_tokens",
	} {
		executeDeleteQuery(db, table, "WHERE user_group_id IN (?)", userGroupsIDs)
	}
	for _, table := range [...]string{"permissions_granted", "permissions_generated", "groups_attempts", "groups_login_prefixes"} {
		executeDeleteQuery(db, table, "WHERE group_id IN (?)", allGroups)
	}
	executeDeleteQuery(db, "groups_groups", "WHERE parent_group_id IN (?)", allGroups)
	executeDeleteQuery(db, "groups_groups", "WHERE child_group_id IN (?)", allGroups)
	executeDeleteQuery(db, "groups_ancestors", "WHERE ancestor_group_id IN (?)", allGroups)
	executeDeleteQuery(db, "groups_ancestors", "WHERE child_group_id IN (?)", allGroups)
	for _, table := range [...]string{"groups_propagate", "groups"} {
		executeDeleteQuery(db, table, "WHERE id IN (?)", allGroups)
	}
	executeDeleteQuery(db, "users", "WHERE group_id IN (?)", userGroupsIDs)
}

func executeDeleteQuery(s *DB, table, condition string, args ...interface{}) {
	mustNotBeError(
		s.Exec(fmt.Sprintf("DELETE %[1]s FROM %[1]s ", QuoteName(table))+condition, args...).Error()) //nolint:gosec
}
