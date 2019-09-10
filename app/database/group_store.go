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

// TeamGroupForTeamItemAndUser returns a composable query for getting a team that
//  1) the given user is a member of
//  2) has `idTeamItem` equal to the given `itemID`.
// If more than one team is found (which should be impossible), the one with the smallest `groups.ID` is returned.
func (s *GroupStore) TeamGroupForTeamItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.idGroupParent = groups.ID AND
				groups_groups.sType`+GroupRelationIsActiveCondition+` AND
				groups_groups.idGroupChild = ?`, user.SelfGroupID).
		Where("groups.idTeamItem = ?", itemID).
		Where("groups.sType = 'Team'").
		Order("groups.ID").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}

// TeamGroupForItemAndUser returns a composable query for getting a team that
//  1) the given user is a member of
//  2) has `idTeamItem` equal to the given `itemID` or one of its ancestors.
// If more than one team is found, the one with the smallest `groups.ID` is returned.
func (s *GroupStore) TeamGroupForItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.idGroupParent = groups.ID AND
				groups_groups.sType`+GroupRelationIsActiveCondition+` AND
				groups_groups.idGroupChild = ?`, user.SelfGroupID).
		Joins(`LEFT JOIN items_ancestors
			ON items_ancestors.idItemAncestor = groups.idTeamItem`).
		Where("groups.sType = 'Team'").
		Where("items_ancestors.idItemChild = ? OR groups.idTeamItem = ?", itemID, itemID).
		Group("groups.ID").
		Order("groups.ID").
		Limit(1)
}

// TeamsMembersForItem returns a composable query for getting all the actual team members for given teamItemID.
// IDs of members' self groups can be fetched as `groups_groups.idGroupChild` while the teams go as `groups`.
func (s *GroupStore) TeamsMembersForItem(groupsToCheck []int64, teamItemID int64) *DB {
	return s.
		Joins(`
			JOIN groups_groups
				ON groups_groups.idGroupParent = groups.ID AND
					groups_groups.sType`+GroupRelationIsActiveCondition).
		Where("groups.sType = 'Team'").
		Where("groups_groups.idGroupChild IN (?)", groupsToCheck).
		Where("groups.idTeamItem = ?", teamItemID)
}
