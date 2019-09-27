package database

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	*DataStore
}

// MatchingUserAncestors returns a composable query of group items matching groups of which the user is member
func (s *GroupItemStore) MatchingUserAncestors(user *User) *DB {
	db := s.GroupAncestors().UserAncestors(user)
	userAncestors := db.SubQuery()
	return s.Joins("JOIN ? AS ancestors ON groups_items.group_id = ancestors.ancestor_group_id", userAncestors)
}

// After is a "listener" that calls GroupItemStore::computeAllAccess() & GroupItemStore::grantCachedAccessWhereNeeded()
func (s *GroupItemStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.computeAllAccess()
	s.grantCachedAccessWhereNeeded()
	return nil
}

func (s *GroupItemStore) removePartialAccess(groupID, itemID int64) {
	mustNotBeError(s.Where("item_id = ? AND group_id = ? AND manager_access = 0", itemID, groupID).
		UpdateColumn(map[string]interface{}{
			"partial_access_since":        nil,
			"cached_partial_access_since": nil,
			"cached_partial_access":       0,
		}).Error())
}

// AccessRightsForItemsVisibleToGroup returns a composable query for getting access rights
// (as full_access, partial_access, grayed_access, access_solutions) and item ids (as item_id)
// for all the items that are visible to the given group.
// Note that the `groupID` can be nil.
func (s *GroupItemStore) AccessRightsForItemsVisibleToGroup(groupID *int64) *DB {
	return s.
		Select(`
			item_id,
			MIN(cached_full_access_since) <= NOW() AS full_access,
			MIN(cached_partial_access_since) <= NOW() AS partial_access,
			MIN(cached_grayed_access_since) <= NOW() AS grayed_access,
			MIN(cached_solutions_access_since) <= NOW() AS access_solutions`).
		Joins(`
			JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.child_group_id = ?)) AS ancestors
			ON ancestors.ancestor_group_id = groups_items.group_id`, groupID).
		Group("groups_items.item_id").
		Having("full_access > 0 OR partial_access > 0 OR grayed_access > 0")
}

// AccessRightsForItemsVisibleToUser returns a composable query for getting access rights
// (as full_access, partial_access, grayed_access, access_solutions) and item ids (as item_id)
// for all the items that are visible to the given user.
func (s *GroupItemStore) AccessRightsForItemsVisibleToUser(user *User) *DB {
	return s.AccessRightsForItemsVisibleToGroup(user.SelfGroupID)
}
