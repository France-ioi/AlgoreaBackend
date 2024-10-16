package database

import (
	"fmt"
	"time"
)

// UserStore implements database operations on `users`.
type UserStore struct {
	*DataStore
}

// ByID returns a composable query for filtering by _table_.group_id.
func (s *UserStore) ByID(id int64) *DB {
	return s.Where(s.tableName+".group_id = ?", id)
}

const deleteWithTrapsBatchSize = 1000

// DeleteTemporaryWithTraps deletes temporary users who don't have sessions or whose sessions expired more than `delay` ago.
// It also removes linked rows in the tables:
//  1. [`filters`, `sessions`, `access_tokens`]
//     having `user_id` = `users.group_id`;
//  2. `answers` having `author_id`/`participant_id` = `users.group_id`;
//  3. [`permissions_granted`, `permissions_generated`, `attempts`]
//     having `group_id` = `users.group_id`;
//  4. [`attempts`, `results`]
//     having `participant_id` = `users.group_id`;
//  5. `groups_groups` having `parent_group_id` or `child_group_id` equal to `users.group_id`;
//  6. `groups_ancestors` having `ancestor_group_id` or `child_group_id` equal to `users.group_id`;
//  7. [`groups_propagate`, `groups`] having `id` equal to `users.group_id`.
func (s *UserStore) DeleteTemporaryWithTraps(delay time.Duration) (err error) {
	defer recoverPanics(&err)

	s.executeBatchesInTransactions(func(store *DataStore) int {
		userScope := store.Users().
			Joins("LEFT JOIN sessions ON sessions.user_id = users.group_id").
			Joins(`
				LEFT JOIN access_tokens ON access_tokens.session_id = sessions.session_id AND
					access_tokens.expires_at > NOW() - INTERVAL ? SECOND`,
				uint64(delay.Round(time.Second)/time.Second)).
			Where("access_tokens.session_id IS NULL").
			Where("temp_user = 1")
		return store.Users().deleteWithTraps(userScope)
	})
	return nil
}

// DeleteWithTraps deletes a given user. It also removes linked rows in the same way as DeleteTemporaryWithTraps.
func (s *UserStore) DeleteWithTraps(user *User) (err error) {
	return s.InTransaction(func(store *DataStore) error {
		deleteOneBatchOfUsers(store.DB, []int64{user.GroupID})
		store.GroupGroups().createNewAncestors()
		return nil
	})
}

// DeleteWithTrapsByScope deletes users matching the given scope.
// It also removes linked rows in the same way as DeleteTemporaryWithTraps.
func (s *UserStore) DeleteWithTrapsByScope(scopeFunc func(store *DataStore) *DB) (err error) {
	defer recoverPanics(&err)

	s.executeBatchesInTransactions(func(store *DataStore) int {
		scope := scopeFunc(store)
		return store.Users().deleteWithTraps(scope)
	})
	return nil
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
// and all the users' stuff.
func (s *UserStore) deleteWithTraps(userScope *DB) int {
	userScope.mustBeInTransaction()

	userIDs := make([]int64, 0, deleteWithTrapsBatchSize)

	mustNotBeError(
		userScope.WithWriteLock().Select("group_id").Limit(deleteWithTrapsBatchSize).
			ScanIntoSlices(&userIDs).Error())

	if len(userIDs) == 0 {
		return 0
	}

	deleteOneBatchOfUsers(userScope, userIDs)
	s.GroupGroups().createNewAncestors()

	return len(userIDs)
}

func deleteOneBatchOfUsers(db *DB, userIDs []int64) {
	db.mustBeInTransaction()

	// we should delete from groups_groups explicitly in order to invoke triggers on groups_groups
	executeDeleteQuery(db, "groups_groups", "WHERE parent_group_id IN (?)", userIDs)
	// deleting from `groups` triggers deletion from
	// `groups_propagate`, `groups_groups`, `groups_ancestors`, `group_pending_requests`, `group_membership_changes`,
	// `permissions_granted`, `permissions_generated", `attempts`, `results`,
	// `users`, `answers`, `filters`, `sessions`, `access_tokens`
	executeDeleteQuery(db, "groups", "WHERE id IN (?)", userIDs)
}

func executeDeleteQuery(s *DB, table, condition string, args ...interface{}) {
	mustNotBeError(
		s.Exec(fmt.Sprintf("DELETE %[1]s FROM %[1]s ", QuoteName(table))+condition, args...).Error())
}
