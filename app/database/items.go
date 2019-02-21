package database

// WhereItemsAreVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsAreVisible(user AuthUser) *DB {
	groupItemsPerms := NewDataStore(newDB(conn.db.New())).GroupItems().
		MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Group("idItem")

	return conn.Joins("JOIN ? as visible ON visible.idItem = items.ID", groupItemsPerms.SubQuery()).
		Where("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0")
}

// JoinsUserAndDefaultItemStrings joins items_strings with the given view twice
// (as default_strings for item's default language and as user_strings for the user's default language)
func (conn *DB) JoinsUserAndDefaultItemStrings(user AuthUser) *DB {
	return conn.
		Joins(
			`LEFT JOIN items_strings default_strings FORCE INDEX (idItem)
         ON default_strings.idItem = items.ID AND default_strings.idLanguage = items.idDefaultLanguage`).
		Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.idItem=items.ID AND user_strings.idLanguage = ?`, user.DefaultLanguageID())
}
