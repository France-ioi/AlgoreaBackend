package database

import (
	"fmt"
)

// UserStore implements database operations on `users`
type UserStore struct {
	*DataStore
}

const deleteWithTrapsBatchSize = 1000

// DeleteTemporaryWithTraps deletes temporary users and all their stuff
func (s *UserStore) DeleteTemporaryWithTraps() (err error) {
	defer recoverPanics(&err)

	s.executeBatchesInTransactions(func(store *DataStore) int {
		userScope := store.Users().
			Joins("LEFT JOIN sessions ON sessions.idUser = users.ID AND sessions.sExpirationDate > NOW()").
			Where("sessions.idUser IS NULL").Where("tempUser = 1")
		return store.Users().deleteWithTraps(userScope)
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
// and all the users' stuff
func (s *UserStore) deleteWithTraps(userScope *DB) int {
	userScope.mustBeInTransaction()

	ids := make([]int64, 0, deleteWithTrapsBatchSize)
	selfGroupsIDs := make([]*int64, 0, deleteWithTrapsBatchSize)  // can be NULL
	ownedGroupsIDs := make([]*int64, 0, deleteWithTrapsBatchSize) // can be NULL

	mustNotBeError(
		userScope.WithWriteLock().Select("ID, idGroupSelf, idGroupOwned").Limit(deleteWithTrapsBatchSize).
			ScanIntoSlices(&ids, &selfGroupsIDs, &ownedGroupsIDs).Error())

	if len(ids) == 0 {
		return 0
	}

	deleteOneBatchOfUsers(userScope, ids, selfGroupsIDs, ownedGroupsIDs)

	return len(ids)
}

func deleteOneBatchOfUsers(db *DB, ids []int64, selfGroupsIDs, ownedGroupsIDs []*int64) {
	db.mustBeInTransaction()

	allGroups := make([]*int64, 0, len(selfGroupsIDs)+len(ownedGroupsIDs))
	allGroups = append(allGroups, selfGroupsIDs...)
	allGroups = append(allGroups, ownedGroupsIDs...)
	for _, table := range [...]string{
		"users_threads", "history_users_threads", "users_answers",
		"users_items", "history_users_items", "filters", "history_filters",
		"sessions", "refresh_tokens",
	} {
		executeDeleteQuery(db, table, "WHERE idUser IN (?)", ids)
	}
	executeDeleteQuery(db, "groups_items_propagate",
		"JOIN groups_items ON groups_items.ID = groups_items_propagate.ID WHERE groups_items.idGroup IN (?)", allGroups)
	for _, table := range [...]string{
		"groups_items", "history_groups_items", "groups_attempts", "history_groups_attempts",
		"groups_login_prefixes", "history_groups_login_prefixes",
	} {
		executeDeleteQuery(db, table, "WHERE idGroup IN (?)", allGroups)
	}
	for _, table := range [...]string{"groups_groups", "history_groups_groups"} {
		executeDeleteQuery(db, table, "WHERE idGroupParent IN (?)", allGroups)
		executeDeleteQuery(db, table, "WHERE idGroupChild IN (?)", allGroups)
	}
	for _, table := range [...]string{"groups_ancestors", "history_groups_ancestors"} {
		executeDeleteQuery(db, table, "WHERE idGroupAncestor IN (?)", allGroups)
		executeDeleteQuery(db, table, "WHERE idGroupChild IN (?)", allGroups)
	}
	for _, table := range [...]string{"groups_propagate", "groups", "history_groups"} {
		executeDeleteQuery(db, table, "WHERE ID IN (?)", allGroups)
	}
	for _, table := range [...]string{"users", "history_users"} {
		executeDeleteQuery(db, table, "WHERE ID IN (?)", ids)
	}
}

func executeDeleteQuery(s *DB, table, condition string, args ...interface{}) {
	mustNotBeError(
		s.Exec(fmt.Sprintf("DELETE %[1]s FROM %[1]s ", table)+condition, args...).Error()) //nolint:gosec
}
