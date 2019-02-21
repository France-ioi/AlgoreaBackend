package database

// WhereItemsVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsVisible(user AuthUser) *DB {
	groupItemsPerms := NewDataStore(NewDB(conn.DB.New())).GroupItems().
		MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Group("idItem")

	return conn.Joins("JOIN ? as visible ON visible.idItem = items.ID", groupItemsPerms.SubQuery()).
		Where("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0")
}
