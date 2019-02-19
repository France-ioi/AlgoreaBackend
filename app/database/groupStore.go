package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// ByID returns a composable query of groups filtered by groupID
func (s *GroupStore) ByID(groupID int64) *DB {
	return s.Where("groups.ID = ?", groupID)
}

// OwnedBy returns a composable query for getting all the groups
// that are descendants of the user's owned group using AuthUser object
func (s *GroupStore) OwnedBy(user AuthUser) *DB {
	return s.Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID").
		Where("groups_ancestors.idGroupAncestor=?", user.OwnedGroupID())
}
