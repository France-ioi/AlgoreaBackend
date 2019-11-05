package database

// WhereItemsAreVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsAreVisible(user *User) *DB {
	visibleItemsPerms := NewDataStore(newDB(conn.db.New())).PermissionsGenerated().AccessRightsForItemsVisibleToUser(user)
	return conn.Joins("JOIN ? as visible ON visible.item_id = items.id", visibleItemsPerms.SubQuery())
}

// JoinsUserAndDefaultItemStrings joins items_strings with the given view twice
// (as default_strings for item's default language and as user_strings for the user's default language)
func (conn *DB) JoinsUserAndDefaultItemStrings(user *User) *DB {
	return conn.
		Joins(
			`LEFT JOIN items_strings default_strings FORCE INDEX (item_id)
         ON default_strings.item_id = items.id AND default_strings.language_id = items.default_language_id`).
		Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.item_id=items.id AND user_strings.language_id = ?`, user.DefaultLanguageID)
}
