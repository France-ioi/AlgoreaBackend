package database

// GroupStore implements database operations on groups
type GroupStore struct {
	*DataStore
}

// OwnedBy returns a composable query for getting all the groups
// that are descendants of the user's owned group using a User object
func (s *GroupStore) OwnedBy(user *User) *DB {
	return s.Joins("JOIN groups_ancestors ON groups_ancestors.group_child_id = groups.id").
		Where("groups_ancestors.group_ancestor_id=?", user.OwnedGroupID)
}

// TeamGroupForTeamItemAndUser returns a composable query for getting a team that
//  1) the given user is a member of
//  2) has `team_item_id` equal to the given `itemID`.
// If more than one team is found (which should be impossible), the one with the smallest `groups.id` is returned.
func (s *GroupStore) TeamGroupForTeamItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.group_parent_id = groups.id AND
				groups_groups.type`+GroupRelationIsActiveCondition+` AND
				groups_groups.group_child_id = ?`, user.SelfGroupID).
		Where("groups.team_item_id = ?", itemID).
		Where("groups.type = 'Team'").
		Order("groups.id").
		Limit(1) // The current API doesn't allow users to join multiple teams working on the same item
}

// TeamGroupForItemAndUser returns a composable query for getting a team that
//  1) the given user is a member of
//  2) has `team_item_id` equal to the given `itemID` or one of its ancestors.
// If more than one team is found, the one with the smallest `groups.id` is returned.
func (s *GroupStore) TeamGroupForItemAndUser(itemID int64, user *User) *DB {
	return s.
		Joins(`JOIN groups_groups
			ON groups_groups.group_parent_id = groups.id AND
				groups_groups.type`+GroupRelationIsActiveCondition+` AND
				groups_groups.group_child_id = ?`, user.SelfGroupID).
		Joins(`LEFT JOIN items_ancestors
			ON items_ancestors.item_ancestor_id = groups.team_item_id`).
		Where("groups.type = 'Team'").
		Where("items_ancestors.item_child_id = ? OR groups.team_item_id = ?", itemID, itemID).
		Group("groups.id").
		Order("groups.id").
		Limit(1)
}

// TeamsMembersForItem returns a composable query for getting all the actual team members for given teamItemID.
// IDs of members' self groups can be fetched as `groups_groups.group_child_id` while the teams go as `groups`.
func (s *GroupStore) TeamsMembersForItem(groupsToCheck []int64, teamItemID int64) *DB {
	return s.
		Joins(`
			JOIN groups_groups
				ON groups_groups.group_parent_id = groups.id AND
					groups_groups.type`+GroupRelationIsActiveCondition).
		Where("groups.type = 'Team'").
		Where("groups_groups.group_child_id IN (?)", groupsToCheck).
		Where("groups.team_item_id = ?", teamItemID)
}
