package database

import (
	"errors"
	"time"
)

// GroupGroupStore implements database operations on `groups_groups`
// (which stores parent-child relationships between groups.
// The relationship is only effective when the 'direct'  type is set.)
type GroupGroupStore struct {
	*DataStore
}

// WhereUserIsMember returns a composable query of direct ancestors of user's self group,
// i.e. groups of which he is a direct member
func (s *GroupGroupStore) WhereUserIsMember(user *User) *DB {
	userSelfGroupID, err := user.SelfGroupID()
	if err != nil {
		_ = s.DB.db.AddError(err)
		return s.DB
	}

	return s.Where("groups_groups.idGroupChild = ? AND groups_groups.sType='direct'", userSelfGroupID)
}

func (s *GroupGroupStore) createNewAncestors() {
	s.DataStore.createNewAncestors("groups", "Group")
}

var ErrRelationCycle = errors.New("a group cannot become an ancestor of itself")

const groupsRelationsLockTimeout = 3 * time.Second

func (s *GroupGroupStore) CreateRelation(parentGroupID, childGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) (err error) {
		mustNotBeError(store.GroupGroups().
			db.Delete(nil, "idGroupChild = ? AND idGroupParent = ?", childGroupID, parentGroupID).Error)

		var rows []interface{}
		mustNotBeError(store.GroupAncestors().
			WithWriteLock().
			Select("ID").
			Where("idGroupChild = ? AND idGroupAncestor = ?", parentGroupID, childGroupID).
			Limit(1).
			Scan(&rows).Error())
		if len(rows) > 0 {
			return ErrRelationCycle
		}

		groupGroupStore := store.GroupGroups()

		mustNotBeError(groupGroupStore.db.Exec(
			"SET @maxIChildOrder = IFNULL((SELECT MAX(iChildOrder) FROM `groups_groups` WHERE `idGroupParent` = ? FOR UPDATE), 0)",
			parentGroupID).Error)

		mustNotBeError(groupGroupStore.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
			store := NewDataStore(db).GroupGroups()
			newID := store.NewID()
			return store.db.Exec(
				"INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) VALUES (?, ?, ?, @maxIChildOrder+1)",
				newID, parentGroupID, childGroupID).Error
		}))

		groupGroupStore.createNewAncestors()
		return nil
	}))
	return
}
