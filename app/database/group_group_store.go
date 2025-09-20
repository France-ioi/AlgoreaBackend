package database

import (
	"errors"
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

// ErrRelationCycle is returned by CreateRelation() if the relation is impossible because it would
// create a cycle in the groups_groups graph.
var ErrRelationCycle = errors.New("a group cannot become an ancestor of itself")

// ParentChild represents a (ParentID, ChildID) pair.
type ParentChild struct {
	ParentID int64
	ChildID  int64
}

// CreateRelation creates a direct relation between two groups.
// On success, it creates new groups ancestors and schedules permissions propagation.
func (s *GroupGroupStore) CreateRelation(parentGroupID, childGroupID int64) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	found, err := s.GroupAncestors().
		// this "FOR SHARE" prevents other sessions from creating this row in groups_ancestors concurrently
		// so there's no need to use a named lock here
		WithSharedWriteLock().
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
	groupGroupStore.createNewAncestors()
	s.ScheduleResultsPropagation()

	return err
}

// CreateRelationsWithoutChecking creates multiple direct relations at once
// without checking for possible cycles in the graph and without deletion of old relations.
// This method is only suitable to create relations with new groups.
// Callers must check that parent groups are not users.
func (s *GroupGroupStore) CreateRelationsWithoutChecking(relations []map[string]interface{}) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	groupGroupStore := s.GroupGroups()
	mustNotBeError(s.InsertMaps(relations))
	groupGroupStore.createNewAncestors()

	return err
}

// DeleteRelation deletes a relation between two groups. It can also delete orphaned groups.
func (s *GroupGroupStore) DeleteRelation(parentGroupID, childGroupID int64, shouldDeleteOrphans bool) (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	// check if parent_group_id is the only parent of child_group_id
	var shouldDeleteChildGroup bool
	shouldDeleteChildGroup, err = s.ActiveGroupGroups().WithExclusiveWriteLock().
		Where("child_group_id = ?", childGroupID).
		Where("parent_group_id != ?", parentGroupID).HasRows()
	mustNotBeError(err)
	shouldDeleteChildGroup = (!shouldDeleteChildGroup) && shouldDeleteOrphans

	if shouldDeleteChildGroup {
		s.GroupGroups().deleteGroupAndOrphanedDescendants(childGroupID)
	} else {
		// delete the relation we are asked to delete (triggers will mark relations for propagation)
		mustNotBeError(s.GroupGroups().Delete("parent_group_id = ? AND child_group_id = ?", parentGroupID, childGroupID).Error())

		permissionsResult := s.PermissionsGranted().
			Delete("origin = 'group_membership' AND source_group_id = ? AND group_id = ?", parentGroupID, childGroupID)
		mustNotBeError(permissionsResult.Error())
		shouldPropagatePermissions := permissionsResult.RowsAffected() > 0

		// recalculate relations
		s.GroupGroups().createNewAncestors()

		if shouldPropagatePermissions {
			s.SchedulePermissionsPropagation()
		}
	}

	return nil
}

// CreateNewAncestors creates ancestors for groups marked as 'todo' in `groups_propagate`.
func (s *GroupGroupStore) CreateNewAncestors() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.createNewAncestors()
	return nil
}

// TeamGroupForTeamItemAndUser returns a composable query for getting a team
//
//	(as groups_groups_active.parent_group_id) that
//	1. the given user is a member of
//	2. has an unexpired attempt with root_item_id = `itemID`.
//
// If more than one team is found (which should be impossible), the one with the smallest `groups.id` is returned.
func (s *GroupGroupStore) TeamGroupForTeamItemAndUser(itemID int64, user *User) *DB {
	return s.
		Where("groups_groups_active.is_team_membership = 1").
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Joins(`
			JOIN attempts ON attempts.participant_id = groups_groups_active.parent_group_id AND
				attempts.root_item_id = ? AND NOW() < attempts.allows_submissions_until`, itemID).
		Order("groups_groups_active.parent_group_id").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}

func (s *GroupGroupStore) deleteGroupAndOrphanedDescendants(groupID int64) {
	// Candidates for deletion are all groups that are descendants of groupID filtered by type
	var candidatesForDeletion []int64
	mustNotBeError(s.ActiveGroupAncestors().WithExclusiveWriteLock().
		Where("is_self = 0").
		Where("ancestor_group_id = ?", groupID).
		Where("child_group_type NOT IN('Base', 'User')").
		Pluck("child_group_id", &candidatesForDeletion).Error())

	// we delete groups_groups linked to groupID here in order to recalculate new ancestors correctly
	groupRelationsDeleted, permissionsDeleted := s.deleteObjectsLinkedToGroups([]int64{groupID})

	// recalculate relations
	if groupRelationsDeleted > 0 {
		s.GroupGroups().createNewAncestors()
	}

	var idsToDelete []int64
	// besides the group with id = groupID, we also want to delete its descendants
	// whose ancestors list consists only of groupID descendants
	// (since they became orphans)
	if len(candidatesForDeletion) > 0 {
		mustNotBeError(s.Groups().WithExclusiveWriteLock().
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
			groupRelationsDeleted, permissionsDeletedForDescendants := s.deleteObjectsLinkedToGroups(idsToDelete)
			permissionsDeleted += permissionsDeletedForDescendants
			if groupRelationsDeleted > 0 {
				s.GroupGroups().createNewAncestors()
			}
		}
	}

	idsToDelete = append(idsToDelete, groupID)
	// cascading deletes from many tables including groups_ancestors, groups_propagate, permission_granted & permissions_generated
	mustNotBeError(s.Groups().Delete("id IN (?)", idsToDelete).Error())

	if permissionsDeleted > 0 {
		s.SchedulePermissionsPropagation()
	}
}

func (s *GroupGroupStore) createNewAncestors() {
	s.DataStore.createNewAncestors("groups", "group")
}

func (s *GroupGroupStore) createRelation(parentGroupID, childGroupID int64) {
	relationMap := map[string]interface{}{
		"parent_group_id": parentGroupID,
		"child_group_id":  childGroupID,
	}
	mustNotBeError(s.GroupGroups().InsertMap(relationMap))
}

func (s *GroupGroupStore) deleteObjectsLinkedToGroups(groupIDs []int64) (groupRelationsDeleted, permissionsDeleted int64) {
	result := s.GroupGroups().Delete("parent_group_id IN(?)", groupIDs)
	mustNotBeError(result.Error())
	groupRelationsDeleted += result.RowsAffected()
	result = s.GroupGroups().Delete("child_group_id IN(?)", groupIDs)
	mustNotBeError(result.Error())
	groupRelationsDeleted += result.RowsAffected()
	mustNotBeError(s.Table("filters").Delete("group_id IN(?)", groupIDs).Error())
	result = s.PermissionsGranted().Delete("source_group_id IN(?)", groupIDs)
	mustNotBeError(result.Error())
	permissionsDeleted = result.RowsAffected()
	return groupRelationsDeleted, permissionsDeleted
}
