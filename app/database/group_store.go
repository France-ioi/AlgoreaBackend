package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// OwnedBy returns a composable query for getting all the groups
// that are descendants of the user's owned group using a User object
func (s *GroupStore) OwnedBy(user *User) *DB {
	return s.Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID").
		Where("groups_ancestors.idGroupAncestor=?", user.OwnedGroupID)
}

// TeamGroupByTeamItemAndUser returns a composable query for getting a team for the current user by the team's main item
func (s *GroupStore) TeamGroupByTeamItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.idGroupParent = groups.ID AND groups_groups.idGroupChild = ?`, user.SelfGroupID).
		Where("groups.idTeamItem = ?", itemID).
		Where("groups.sType = 'Team'").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}

// TeamGroupByItemAndUser returns a composable query for getting a team for the current user by one of team's items
func (s *GroupStore) TeamGroupByItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.idGroupParent = groups.ID AND groups_groups.idGroupChild = ?`, user.SelfGroupID).
		Joins(`LEFT JOIN items_ancestors
			ON items_ancestors.idItemAncestor = groups.idTeamItem`).
		Where("groups.sType = 'Team'").
		Where("items_ancestors.idItemChild = ? OR groups.idTeamItem = ?", itemID, itemID).
		Group("groups.ID").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}
