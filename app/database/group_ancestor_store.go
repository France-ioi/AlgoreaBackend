package database

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
	*DataStore
}

const groupsAncestorsActive = "groups_ancestors_active"

// UserAncestors returns a composable query of ancestors of user's self group, i.e. groups of which he is a member
func (s *GroupAncestorStore) UserAncestors(user *User) *DB {
	result := s.Where(QuoteName(s.tableName)+".child_group_id = ?", user.GroupID)
	if s.tableName != groupsAncestorsActive {
		result = result.Where("NOW() < " + QuoteName(s.tableName) + ".expires_at")
	}
	return result
}

// ManagedByUser returns a composable query for getting all the groups_ancestors rows for groups
// that are descendants of groups managed by the user
func (s *GroupAncestorStore) ManagedByUser(user *User) *DB {
	result := s.
		Joins(`
			JOIN group_managers
				ON group_managers.group_id = `+QuoteName(s.tableName)+`.ancestor_group_id AND
					group_managers.manager_id = ?`, user.GroupID)
	if s.tableName != groupsAncestorsActive {
		result = result.Where("NOW() < " + QuoteName(s.tableName) + ".expires_at")
	}
	return result
}
