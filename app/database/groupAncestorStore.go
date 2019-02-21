package database

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
	*DataStore
}

// UserAncestors returns a composable query of ancestors of user's self group, i.e. groups of which he is a member
func (s *GroupAncestorStore) UserAncestors(user AuthUser) *DB {
	return s.Where("idGroupChild = ?", user.SelfGroupID())
}

// OwnedByUser returns a composable query for getting all the groups_ancestors rows for groups
// that are descendants of the user's owned group using AuthUser object
func (s *GroupAncestorStore) OwnedByUser(user AuthUser) *DB {
	return s.Where("idGroupAncestor=?", user.OwnedGroupID())
}
