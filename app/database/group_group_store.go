package database

import (
	"errors"
	"time"
)

// GroupGroupStore implements database operations on `groups_groups`
// (which stores parent-child relationships between groups).
type GroupGroupStore struct {
	*DataStore
}

// WhereUserIsMember returns a composable query of direct ancestors (parents) of user's self group,
// i.e. groups of which he is a direct member.
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

// CreateRelation creates a direct relation between two groups.
func (s *GroupGroupStore) CreateRelation(parentGroupID, childGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithGroupsRelationsLock(func(s *DataStore) (err error) {
		found, err := s.GroupAncestors().
			WithWriteLock().
			// do not allow cycles even via expired relations
			Where("child_group_id = ? AND ancestor_group_id = ?", parentGroupID, childGroupID).
			HasRows()
		mustNotBeError(err)
		if found {
			return ErrRelationCycle
		}

		groupGroupStore := s.GroupGroups()
		mustNotBeError(groupGroupStore.Delete("child_group_id = ? AND parent_group_id = ?", childGroupID, parentGroupID).Error())
		mustNotBeError(s.GroupPendingRequests().Delete("group_id = ? AND member_id = ?", parentGroupID, childGroupID).Error())

		groupGroupStore.createRelation(parentGroupID, childGroupID)
		s.ScheduleGroupsAncestorsPropagation()
		s.ScheduleResultsPropagation()
		return nil
	}))
	return err
}

func (s *GroupGroupStore) createRelation(parentGroupID, childGroupID int64) {
	relationMap := map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
	}
	mustNotBeError(s.GroupGroups().InsertMap(relationMap))
}

// CreateRelationsWithoutChecking creates multiple direct relations at once
// without checking for possible cycles in the graph and without deletion of old relations.
// This method is only suitable to create relations with new groups.
// Callers must check that parent groups are not users.
func (s *GroupGroupStore) CreateRelationsWithoutChecking(relations []map[string]interface{}) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithGroupsRelationsLock(func(s *DataStore) (err error) {
		mustNotBeError(s.InsertMaps(relations))
		s.ScheduleGroupsAncestorsPropagation()
		return nil
	}))
	return err
}

// DeleteRelation deletes a relation between two groups. It can also delete orphaned groups.
func (s *GroupGroupStore) DeleteRelation(parentGroupID, childGroupID int64, shouldDeleteOrphans bool) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	mustNotBeError(s.WithGroupsRelationsLock(func(s *DataStore) error {
		// check if parent_group_id is the only parent of child_group_id
		var shouldDeleteChildGroup bool
		shouldDeleteChildGroup, err = s.ActiveGroupGroups().WithWriteLock().
			Where("child_group_id = ?", childGroupID).
			Where("parent_group_id != ?", parentGroupID).HasRows()
		mustNotBeError(err)
		shouldDeleteChildGroup = (!shouldDeleteChildGroup) && shouldDeleteOrphans

		if shouldDeleteChildGroup {
			s.GroupGroups().deleteGroupAndOrphanedDescendants(childGroupID)
		} else {
			// delete the relation we are asked to delete (triggers will delete a lot from groups_ancestors and mark relations for propagation)
			mustNotBeError(s.GroupGroups().Delete("parent_group_id = ? AND child_group_id = ?", parentGroupID, childGroupID).Error())

			permissionsResult := s.PermissionsGranted().
				Delete("origin = 'group_membership' AND source_group_id = ? AND group_id = ?", parentGroupID, childGroupID)
			mustNotBeError(permissionsResult.Error())
			shouldPropagatePermissions := permissionsResult.RowsAffected() > 0

			// recalculate relations
			s.ScheduleGroupsAncestorsPropagation()

			if shouldPropagatePermissions {
				s.SchedulePermissionsPropagation()
			}
		}

		return nil
	}))
	return nil
}

func (s *GroupGroupStore) deleteGroupAndOrphanedDescendants(groupID int64) {
	// Candidates for deletion are all groups that are descendants of groupID filtered by type
	var candidatesForDeletion []int64
	mustNotBeError(s.Groups().WithWriteLock().
		Joins(`
			JOIN groups_ancestors_active AS ancestors ON
				ancestors.child_group_id = groups.id AND
				ancestors.is_self = 0 AND
				ancestors.ancestor_group_id = ?`, groupID).
		Where("groups.type NOT IN('Base', 'User')").
		Pluck("groups.id", &candidatesForDeletion).Error())

	// we delete groups_groups linked to groupID here in order to recalculate new ancestors correctly
	mustNotBeError(s.deleteObjectsLinkedToGroups([]int64{groupID}).Error())

	// recalculate relations
	// It seems (to be verified), that this propagation has to be called right now
	// because it's result is used in the following steps.
	s.GroupGroups().createNewAncestors()

	var idsToDelete []int64
	// besides the group with id = groupID, we also want to delete its descendants
	// whose ancestors list consists only of groupID descendants
	// (since they became orphans)
	if len(candidatesForDeletion) > 0 {
		mustNotBeError(s.Groups().WithWriteLock().
			Joins(`
				LEFT JOIN(
					SELECT groups_ancestors_active.child_group_id
					FROM groups_ancestors_active
					WHERE
						groups_ancestors_active.ancestor_group_id NOT IN(?) AND
						groups_ancestors_active.child_group_id IN(?) AND
						groups_ancestors_active.is_self = 0
					GROUP BY groups_ancestors_active.child_group_id
					FOR UPDATE
				) AS ancestors
				ON ancestors.child_group_id = groups.id`, candidatesForDeletion, candidatesForDeletion).
			Where("groups.id IN (?)", candidatesForDeletion).
			Where("ancestors.child_group_id IS NULL").
			Pluck("groups.id", &idsToDelete).Error())

		if len(idsToDelete) > 0 {
			deleteResult := s.deleteObjectsLinkedToGroups(idsToDelete)
			mustNotBeError(deleteResult.Error())
			if deleteResult.RowsAffected() > 0 {
				s.ScheduleGroupsAncestorsPropagation()
			}
		}
	}

	idsToDelete = append(idsToDelete, groupID)
	// cascading deletes from many tables including groups_ancestors, groups_propagate, permission_granted & permissions_generated
	mustNotBeError(s.Groups().Delete("id IN (?)", idsToDelete).Error())

	s.SchedulePermissionsPropagation()
}

func (s *GroupGroupStore) deleteObjectsLinkedToGroups(groupIDs []int64) *DB {
	return s.Exec(`
		DELETE group_children, group_parents, filters
		FROM `+"`groups`"+`
		LEFT JOIN groups_groups AS group_children
			ON group_children.parent_group_id = groups.id
		LEFT JOIN groups_groups AS group_parents
			ON group_parents.child_group_id = groups.id
		LEFT JOIN filters
			ON filters.group_id = groups.id
		WHERE groups.id IN(?)`, groupIDs)
}

// WithGroupsRelationsLock wraps the given function in GET_LOCK/RELEASE_LOCK
// specific for modifying relations between groups.
func (s *GroupGroupStore) WithGroupsRelationsLock(txFunc func(*DataStore) error) error {
	return s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, txFunc)
}

// RemoveMembersOfGroup removes members of a group.
func (s *GroupGroupStore) RemoveMembersOfGroup(groupID int64, memberIDs []int64) {
	err := s.
		Where("groups_groups.parent_group_id = ?", groupID).
		Where("groups_groups.child_group_id IN (?)", memberIDs).
		Delete().
		Error()
	mustNotBeError(err)
}
