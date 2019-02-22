package database

// GroupGroupStore implements database operations on `groups_groups`
// (which stores parent-child relationships between groups.
// The relationship is only effective when the 'direct'  type is set.)
type GroupGroupStore struct {
	*DataStore
}

// WhereUserIsMember returns a composable query of direct ancestors of user's self group,
// i.e. groups of which he is a direct member
func (s *GroupGroupStore) WhereUserIsMember(user AuthUser) *DB {
	return s.Where("groups_groups.idGroupChild = ? AND groups_groups.sType='direct'", user.SelfGroupID())
}
