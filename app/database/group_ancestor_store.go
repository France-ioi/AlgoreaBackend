package database

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
	*DataStore
}

// UserAncestors returns a composable query of ancestors of user's self group, i.e. groups of which he is a member
func (s *GroupAncestorStore) UserAncestors(user *User) *DB {
	return s.Where("groups_ancestors.group_child_id = ?", user.SelfGroupID)
}

// OwnedByUser returns a composable query for getting all the groups_ancestors rows for groups
// that are descendants of the user's owned group using a User object
func (s *GroupAncestorStore) OwnedByUser(user *User) *DB {
	return s.Where("groups_ancestors.group_ancestor_id=?", user.OwnedGroupID)
}
