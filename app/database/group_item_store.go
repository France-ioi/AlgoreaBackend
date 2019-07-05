package database

// GroupItemStore implements database operations on `groups_items`
type GroupItemStore struct {
	*DataStore
}

// MatchingUserAncestors returns a composable query of group items matching groups of which the user is member
func (s *GroupItemStore) MatchingUserAncestors(user *User) *DB {
	db := s.GroupAncestors().UserAncestors(user)
	if db.Error() != nil {
		return db
	}
	userAncestors := db.SubQuery()
	return s.Joins("JOIN ? AS ancestors ON groups_items.idGroup = ancestors.idGroupAncestor", userAncestors)
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
	mustNotBeError(s.Where("idItem = ? AND idGroup = ? AND bManagerAccess = 0", itemID, groupID).
		UpdateColumn(map[string]interface{}{
			"sPartialAccessDate":       nil,
			"sCachedPartialAccessDate": nil,
			"bCachedPartialAccess":     0,
		}).Error())
}

// AccessRightsForItemsVisibleToGroup returns a composable query for getting access rights
// (as fullAccess, partialAccess, grayedAccess, accessSolutions) and item IDs (as idItem)
// for all the items that are visible to the given group.
func (s *GroupItemStore) AccessRightsForItemsVisibleToGroup(groupID int64) *DB {
	return s.
		Select(`
			idItem,
			MIN(sCachedFullAccessDate) <= NOW() AS fullAccess,
			MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess,
			MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess,
			MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions`).
		Joins(`
			JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors
			ON ancestors.idGroupAncestor = groups_items.idGroup`, groupID).
		Group("groups_items.idItem").
		Having("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0")
}

// AccessRightsForItemsVisibleToUser returns a composable query for getting access rights
// (as fullAccess, partialAccess, grayedAccess, accessSolutions) and item IDs (as idItem)
// for all the items that are visible to the given user.
func (s *GroupItemStore) AccessRightsForItemsVisibleToUser(user *User) *DB {
	userSelfGroupID, err := user.SelfGroupID()
	if err != nil {
		_ = s.DB.db.AddError(err)
		return s.DB
	}

	return s.AccessRightsForItemsVisibleToGroup(userSelfGroupID)
}
