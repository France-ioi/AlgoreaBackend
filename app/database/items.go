package database

// WhereItemsAreVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsAreVisible(user *User) *DB {
	visibleItemsPerms := NewDataStore(newDB(conn.db.New())).GroupItems().AccessRightsForItemsVisibleToUser(user)

	if visibleItemsPerms.Error() != nil {
		return visibleItemsPerms
	}
	return conn.Joins("JOIN ? as visible ON visible.idItem = items.ID", visibleItemsPerms.SubQuery())
}

// JoinsUserAndDefaultItemStrings joins items_strings with the given view twice
// (as default_strings for item's default language and as user_strings for the user's default language)
func (conn *DB) JoinsUserAndDefaultItemStrings(user *User) *DB {
	userDefaultLanguageID, err := user.DefaultLanguageID()
	if err != nil {
		_ = conn.db.AddError(err)
		return conn
	}
	return conn.
		Joins(
			`LEFT JOIN items_strings default_strings FORCE INDEX (idItem)
         ON default_strings.idItem = items.ID AND default_strings.idLanguage = items.idDefaultLanguage`).
		Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.idItem=items.ID AND user_strings.idLanguage = ?`, userDefaultLanguageID)
}
