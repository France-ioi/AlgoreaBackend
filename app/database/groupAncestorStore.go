package database

// GroupAncestorStore implements database operations on `groups_ancestors` (which is a precomputed cache over groups_groups)
type GroupAncestorStore struct {
	*DataStore
}

// All creates a composable query without filtering
func (s *GroupAncestorStore) All() DB {
	return s.table("groups_ancestors")
}

// UserAncestors returns a composable query of ancestors of user's self group, i.e. groups of which he is a member
func (s *GroupAncestorStore) UserAncestors(user AuthUser) DB {
	return s.All().Where("idGroupChild = ?", user.SelfGroupID())
}

// OwnedByUser returns a composable query for getting all the groups_ancestors rows for groups
// that are descendants of the user's owned group using AuthUser object
func (s *GroupAncestorStore) OwnedByUser(user AuthUser) DB {
	return s.All().Where("idGroupAncestor=?", user.OwnedGroupID())
}

// OwnedByUserID returns a composable query for getting all the groups_ancestors rows for groups
// that are descendants of the user's owned group using ID
func (s *GroupAncestorStore) OwnedByUserID(ownerUserID int64) DB {
	return s.All().
		Joins("JOIN users ON users.idGroupOwned=groups_ancestors.idGroupAncestor").
		Where("users.ID=?", ownerUserID)
}

// KeepUsersThatAreDescendantsOf joins `groups_ancestors` on idGroupAncestor=groupID & idGroupChild=users.idGroupSelf
func (s *GroupAncestorStore) KeepUsersThatAreDescendantsOf(groupID int64, db DB) DB {
	return db.
		Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild=users.idGroupSelf").
		Where("groups_ancestors.idGroupAncestor = ?", groupID)
}
