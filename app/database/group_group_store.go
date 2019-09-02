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

// WhereUserIsMember returns a composable query of direct ancestors (parents) of user's self group,
// i.e. groups of which he is a direct member
func (s *GroupGroupStore) WhereUserIsMember(user *User) *DB {
	return s.Where("groups_groups.idGroupChild = ?", user.SelfGroupID).
		WhereGroupRelationIsActive()
}

func (s *GroupGroupStore) createNewAncestors() {
	s.DataStore.createNewAncestors("groups", "Group")
}

// ErrRelationCycle is returned by CreateRelation() if the relation is impossible because it would
// create a cycle in the groups_groups graph.
var ErrRelationCycle = errors.New("a group cannot become an ancestor of itself")

const groupsRelationsLockTimeout = 3 * time.Second

// ParentChild represents a (ParentID, ChildID) pair
type ParentChild struct {
	ParentID int64
	ChildID  int64
}

// CreateRelation creates a direct relation between two groups
func (s *GroupGroupStore) CreateRelation(parentGroupID, childGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) (err error) {
		mustNotBeError(store.GroupGroups().Delete("idGroupChild = ? AND idGroupParent = ?", childGroupID, parentGroupID).Error())

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
		groupGroupStore.createRelation(parentGroupID, childGroupID)
		groupGroupStore.createNewAncestors()
		return nil
	}))
	return err
}

func (s *GroupGroupStore) createRelation(parentGroupID, childGroupID int64) {
	s.mustBeInTransaction()
	mustNotBeError(s.db.Exec(
		"SET @maxIChildOrder = IFNULL((SELECT MAX(iChildOrder) FROM `groups_groups` WHERE `idGroupParent` = ? FOR UPDATE), 0)",
		parentGroupID).Error)

	mustNotBeError(s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		store := NewDataStore(db).GroupGroups()
		newID := store.NewID()
		return store.db.Exec(
			"INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) VALUES (?, ?, ?, @maxIChildOrder+1)",
			newID, parentGroupID, childGroupID).Error
	}))
}

// CreateRelationsWithoutChecking creates multiple direct relations between group pairs at once
// without checking for possible cycles in the graph and without deletion of old relations.
// This method is only suitable to create relations with new groups.
func (s *GroupGroupStore) CreateRelationsWithoutChecking(pairs []ParentChild) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) (err error) {
		groupGroupStore := store.GroupGroups()
		for _, pair := range pairs {
			groupGroupStore.createRelation(pair.ParentID, pair.ChildID)
		}
		groupGroupStore.createNewAncestors()
		return nil
	}))
	return err
}

// ErrGroupBecomesOrphan is to be returned if a group is going to become an orphan
// after the relation is deleted. DeleteRelation() returns this error to inform
// the caller that a confirmation is needed (shouldDeleteOrphans should be true).
var ErrGroupBecomesOrphan = errors.New("a group cannot become an orphan")

// DeleteRelation deletes a relation between two groups. It can also delete orphaned groups.
func (s *GroupGroupStore) DeleteRelation(parentGroupID, childGroupID int64, shouldDeleteOrphans bool) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) error {
		// check if idGroupParent is the only parent of idGroupChild
		shouldDeleteChildGroup := false
		var result []interface{}
		mustNotBeError(s.GroupGroups().WithWriteLock().
			Select("1").
			Where("idGroupChild = ?", childGroupID).
			Where("idGroupParent != ?", parentGroupID).
			Limit(1).Scan(&result).Error())
		if len(result) == 0 {
			shouldDeleteChildGroup = true
			if !shouldDeleteOrphans {
				return ErrGroupBecomesOrphan
			}
		}

		var candidatesForDeletion []int64
		if shouldDeleteChildGroup {
			// Candidates for deletion are all groups that are descendants of childGroupID filtered by sType
			mustNotBeError(s.Groups().WithWriteLock().
				Joins(`
					JOIN groups_ancestors AS ancestors ON
						ancestors.idGroupChild = groups.ID AND
						ancestors.bIsSelf = 0 AND
						ancestors.idGroupAncestor = ?`, childGroupID).
				Where("groups.sType NOT IN('Base', 'UserAdmin', 'UserSelf')").
				Pluck("groups.ID", &candidatesForDeletion).Error())
		}

		// triggers delete from groups_ancestors (all except self-links),
		// groups_propagate & groups_items_propagate as well,
		// but the `before_delete_groups_groups` trigger inserts into groups_propagate again :(
		const deleteGroupsQuery = `
			DELETE groups, group_children, group_parents, groups_attempts,
						 groups_items, groups_login_prefixes, filters
			FROM groups
			LEFT JOIN groups_groups AS group_children
				ON group_children.idGroupParent = groups.ID
			LEFT JOIN groups_groups AS group_parents
				ON group_parents.idGroupChild = groups.ID
			LEFT JOIN groups_attempts
				ON groups_attempts.idGroup = groups.ID
			LEFT JOIN groups_items
				ON groups_items.idGroup = groups.ID
			LEFT JOIN groups_login_prefixes
				ON groups_login_prefixes.idGroup = groups.ID
			LEFT JOIN filters
				ON filters.idGroup = groups.ID
			WHERE groups.ID IN(?)`

		// delete the relation we are asked to delete (triggers will delete a lot from groups_ancestors and mark relations for propagation)
		mustNotBeError(s.GroupGroups().Delete("idGroupParent = ? AND idGroupChild = ?", parentGroupID, childGroupID).Error())

		if shouldDeleteChildGroup {
			// we delete the orphan here in order to recalculate new ancestors correctly
			mustNotBeError(s.db.Exec(deleteGroupsQuery, []int64{childGroupID}).Error)
		}
		// recalculate relations
		s.GroupGroups().createNewAncestors()

		if shouldDeleteChildGroup {
			var idsToDelete []int64
			// besides the group with ID = childGroupID, we also want to delete its descendants
			// whose ancestors list consists only of childGroupID descendants
			// (since they would become orphans)
			if len(candidatesForDeletion) > 0 {
				mustNotBeError(s.Groups().WithWriteLock().
					Joins(`
						LEFT JOIN(
							SELECT groups_ancestors.idGroupChild
							FROM groups_ancestors
							WHERE
								groups_ancestors.idGroupAncestor NOT IN(?) AND
								groups_ancestors.idGroupChild IN(?) AND
								groups_ancestors.bIsSelf = 0
							GROUP BY groups_ancestors.idGroupChild
							FOR UPDATE
						) AS ancestors
						ON ancestors.idGroupChild = groups.ID`, candidatesForDeletion, candidatesForDeletion).
					Where("groups.ID IN (?)", candidatesForDeletion).
					Where("ancestors.idGroupChild IS NULL").
					Pluck("groups.ID", &idsToDelete).Error())

				if len(idsToDelete) > 0 {
					deleteResult := s.db.Exec(deleteGroupsQuery, idsToDelete)
					mustNotBeError(deleteResult.Error)
					if deleteResult.RowsAffected > 0 {
						s.GroupGroups().createNewAncestors()
					}
				}
			}

			idsToDelete = append(idsToDelete, childGroupID)
			// delete self relations of the removed groups
			mustNotBeError(s.GroupAncestors().Delete("idGroupAncestor IN (?)", idsToDelete).Error())
			// delete removed groups from groups_propagate
			mustNotBeError(s.Table("groups_propagate").Delete("ID IN (?)", idsToDelete).Error())
		}
		return nil
	}))
	return nil
}
