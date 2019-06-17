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
