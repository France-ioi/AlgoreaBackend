package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// OwnedBy returns a composable query for getting all the groups
// that are descendants of the user's owned group using a User object
func (s *GroupStore) OwnedBy(user *User) *DB {
	userOwnedGroupID, err := user.OwnedGroupID()
	if err != nil {
		_ = s.DB.db.AddError(err)
		return s.DB
	}
	return s.Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID").
		Where("groups_ancestors.idGroupAncestor=?", userOwnedGroupID)
}

// TeamGroupByItemAndUser returns a composable query for getting a team on some item for the current user
func (s *GroupStore) TeamGroupByItemAndUser(itemID int64, user *User) *DB {
	selfGroupID, err := user.SelfGroupID()
	if err != nil {
		_ = s.DB.db.AddError(err)
		return s.DB
	}

	return s.
		Joins(`JOIN groups_groups
				ON groups_groups.idGroupParent = groups.ID AND groups_groups.idGroupChild = ?`, selfGroupID).
		Where("groups.idTeamItem = ?", itemID).
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}
