package database

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
	*DataStore
}

const groupsAncestorsActive = "groups_ancestors_active"

// ManagedByUser returns a composable query for getting all the groups_ancestors rows
// linking manager groups (as ancestor_group_id) to managed groups (as child_group_id)
// where the manager groups are ancestors of the given user.
// Basically the groups_ancestors.child_group_id are the groups the user can manage.
//
// The result may contain duplicated
// `groups_ancestors.ancestor_group_id`-`groups_ancestors.child_group_id` pairs since
// there can be different paths to a managed group through the `group_managers` table and
// the group ancestry graph.
func (s *GroupAncestorStore) ManagedByUser(user *User) *DB {
	return s.ManagedByGroup(user.GroupID)
}

// ManagedByGroup returns a composable query for getting all the groups_ancestors rows
// linking manager groups (as ancestor_group_id) to managed groups (as child_group_id)
// where the manager groups are ancestors of the given group.
// Basically the groups_ancestors.child_group_id are the groups the given group can manage.
//
// The result may contain duplicated
// `groups_ancestors.ancestor_group_id`-`groups_ancestors.child_group_id` pairs since
// there can be different paths to a managed group through the `group_managers` table and
// the group ancestry graph.
func (s *GroupAncestorStore) ManagedByGroup(groupID int64) *DB {
	result := s.
		Joins(`
			JOIN group_managers
				ON group_managers.group_id = `+QuoteName(s.tableName)+`.ancestor_group_id`).
		Joins(`
			JOIN groups_ancestors_active AS group_ancestors
				ON group_ancestors.ancestor_group_id = group_managers.manager_id AND
					group_ancestors.child_group_id = ?`, groupID)
	if s.tableName != groupsAncestorsActive {
		result = result.Where("NOW() < " + QuoteName(s.tableName) + ".expires_at")
	}
	return result
}
