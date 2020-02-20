package database

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// GroupGroupStore implements database operations on `groups_groups`
// (which stores parent-child relationships between groups).
type GroupGroupStore struct {
	*DataStore
}

// WhereUserIsMember returns a composable query of direct ancestors (parents) of user's self group,
// i.e. groups of which he is a direct member
func (s *GroupGroupStore) WhereUserIsMember(user *User) *DB {
	result := s.Where(QuoteName(s.tableName)+".child_group_id = ?", user.GroupID)
	if s.tableName != "groups_groups_active" {
		result = result.WhereGroupRelationIsActual()
	}
	return result
}

func (s *GroupGroupStore) createNewAncestors() {
	s.DataStore.createNewAncestors("groups", "group")
}

// ErrRelationCycle is returned by CreateRelation() if the relation is impossible because it would
// create a cycle in the groups_groups graph.
var ErrRelationCycle = errors.New("a group cannot become an ancestor of itself")

const groupsRelationsLockTimeout = 3 * time.Second

// ParentChild represents a (ParentID, ChildID) pair.
type ParentChild struct {
	ParentID int64
	ChildID  int64
}

// CreateRelation creates a direct relation between two groups
func (s *GroupGroupStore) CreateRelation(parentGroupID, childGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) (err error) {
		mustNotBeError(store.GroupGroups().Delete("child_group_id = ? AND parent_group_id = ?", childGroupID, parentGroupID).Error())
		mustNotBeError(store.GroupPendingRequests().Delete("group_id = ? AND member_id = ?", parentGroupID, childGroupID).Error())

		found, err := store.GroupAncestors().
			WithWriteLock().
			// do not allow cycles even via expired relations
			Where("child_group_id = ? AND ancestor_group_id = ?", parentGroupID, childGroupID).
			HasRows()
		mustNotBeError(err)
		if found {
			return ErrRelationCycle
		}

		groupGroupStore := store.GroupGroups()
		groupGroupStore.createRelation(parentGroupID, childGroupID)
		groupGroupStore.createNewAncestors()
		return store.Attempts().ComputeAllAttempts()
	}))
	return err
}

func (s *GroupGroupStore) createRelation(parentGroupID, childGroupID int64) {
	s.mustBeInTransaction()
	mustNotBeError(s.db.Exec(
		"SET @maxIChildOrder = IFNULL((SELECT MAX(child_order) FROM `groups_groups` WHERE `parent_group_id` = ? FOR UPDATE), 0)",
		parentGroupID).Error)

	relationMap := map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
		"child_order":     gorm.Expr("@maxIChildOrder+1"),
	}
	mustNotBeError(s.GroupGroups().InsertMap(relationMap))
}

// CreateRelationsWithoutChecking creates multiple direct relations at once
// without checking for possible cycles in the graph and without deletion of old relations.
// This method is only suitable to create relations with new groups.
func (s *GroupGroupStore) CreateRelationsWithoutChecking(relations []map[string]interface{}) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	parentRelations := make(map[int64][]map[string]interface{})
	for _, relation := range relations {
		parentRelations[relation["parent_group_id"].(int64)] = append(parentRelations[relation["parent_group_id"].(int64)], relation)
	}
	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(store *DataStore) (err error) {
		groupGroupStore := store.GroupGroups()
		for parentID, groupedRelations := range parentRelations {
			for index := range groupedRelations {
				groupedRelations[index]["child_order"] = gorm.Expr("@maxIChildOrder+?", index+1)
			}
			mustNotBeError(s.Exec(
				"SET @maxIChildOrder = IFNULL((SELECT MAX(child_order) FROM `groups_groups` WHERE `parent_group_id` = ? FOR UPDATE), 0)",
				parentID).Error())
			mustNotBeError(s.InsertMaps(groupedRelations))
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
		// check if parent_group_id is the only parent of child_group_id
		var shouldDeleteChildGroup bool
		shouldDeleteChildGroup, err = s.GroupGroups().WithWriteLock().
			Where("child_group_id = ?", childGroupID).
			Where("parent_group_id != ?", parentGroupID).HasRows()
		mustNotBeError(err)
		shouldDeleteChildGroup = !shouldDeleteChildGroup
		if shouldDeleteChildGroup && !shouldDeleteOrphans {
			return ErrGroupBecomesOrphan
		}

		var candidatesForDeletion []int64
		if shouldDeleteChildGroup {
			// Candidates for deletion are all groups that are descendants of childGroupID filtered by type
			mustNotBeError(s.Groups().WithWriteLock().
				Joins(`
					JOIN groups_ancestors AS ancestors ON
						ancestors.child_group_id = groups.id AND
						ancestors.is_self = 0 AND
						ancestors.ancestor_group_id = ?`, childGroupID).
				Where("groups.type NOT IN('Base', 'User')").
				Pluck("groups.id", &candidatesForDeletion).Error())
		}

		const deleteGroupsQuery = `
			DELETE group_children, group_parents, attempts, filters
			FROM ` + "`groups`" + `
			LEFT JOIN groups_groups AS group_children
				ON group_children.parent_group_id = groups.id
			LEFT JOIN groups_groups AS group_parents
				ON group_parents.child_group_id = groups.id
			LEFT JOIN attempts
				ON attempts.group_id = groups.id
			LEFT JOIN filters
				ON filters.group_id = groups.id
			WHERE groups.id IN(?)`

		// delete the relation we are asked to delete (triggers will delete a lot from groups_ancestors and mark relations for propagation)
		mustNotBeError(s.GroupGroups().Delete("parent_group_id = ? AND child_group_id = ?", parentGroupID, childGroupID).Error())

		var shouldPropagatePermissions bool

		if shouldDeleteChildGroup {
			// we delete the orphan here in order to recalculate new ancestors correctly
			// (no need to delete permissions here since we have a cascade delete in the DB)
			mustNotBeError(s.db.Exec(deleteGroupsQuery, []int64{childGroupID}).Error)
			shouldPropagatePermissions = true
		} else {
			permissionsResult := store.PermissionsGranted().
				Delete("origin = 'group_membership' AND source_group_id = ? AND group_id = ?", parentGroupID, childGroupID)
			mustNotBeError(permissionsResult.Error())
			shouldPropagatePermissions = permissionsResult.RowsAffected() > 0
		}

		// recalculate relations
		s.GroupGroups().createNewAncestors()

		if shouldDeleteChildGroup {
			var idsToDelete []int64
			// besides the group with id = childGroupID, we also want to delete its descendants
			// whose ancestors list consists only of childGroupID descendants
			// (since they would become orphans)
			if len(candidatesForDeletion) > 0 {
				mustNotBeError(s.Groups().WithWriteLock().
					Joins(`
						LEFT JOIN(
							SELECT groups_ancestors.child_group_id
							FROM groups_ancestors
							WHERE
								groups_ancestors.ancestor_group_id NOT IN(?) AND
								groups_ancestors.child_group_id IN(?) AND
								groups_ancestors.is_self = 0
							GROUP BY groups_ancestors.child_group_id
							FOR UPDATE
						) AS ancestors
						ON ancestors.child_group_id = groups.id`, candidatesForDeletion, candidatesForDeletion).
					Where("groups.id IN (?)", candidatesForDeletion).
					Where("ancestors.child_group_id IS NULL").
					Pluck("groups.id", &idsToDelete).Error())

				if len(idsToDelete) > 0 {
					deleteResult := s.Exec(deleteGroupsQuery, idsToDelete)
					mustNotBeError(deleteResult.Error())
					if deleteResult.RowsAffected() > 0 {
						s.GroupGroups().createNewAncestors()
					}
				}
			}

			idsToDelete = append(idsToDelete, childGroupID)
			// triggers/cascading delete from groups_ancestors and groups_propagate,
			mustNotBeError(s.Groups().Delete("id IN (?)", idsToDelete).Error())
		}

		if shouldPropagatePermissions {
			s.PermissionsGranted().computeAllAccess()
		}

		return nil
	}))
	return nil
}

// After is a "listener" that calls GroupGroupStore::createNewAncestors()
func (s *GroupGroupStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.createNewAncestors()
	return nil
}
