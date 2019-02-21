package database

// GroupGroupStore implements database operations on `groups_groups`
// (which stores parent-child relationships between groups.
// The relationship is only effective when the 'direct'  type is set.)
type GroupGroupStore struct {
	*DataStore
}

// All creates a composable query without filtering
func (s *GroupGroupStore) All() *DB {
	return s.Table("groups_groups")
}

// WhereUserIsMember returns a composable query of direct ancestors of user's self group,
// i.e. groups of which he is a direct member
func (s *GroupGroupStore) WhereUserIsMember(user AuthUser) *DB {
	return s.All().Where("idGroupChild = ? AND sType='direct'", user.SelfGroupID())
}
