package database

// WhereUserHasViewPermissionOnItems returns a subview of the items
// on that the given user has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereUserHasViewPermissionOnItems(user *User, viewPermission string) *DB {
	return conn.WhereUserHasPermissionOnItems(user, "view", viewPermission)
}

// WhereUserHasPermissionOnItems returns a subview of the items
// on that the given user has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereUserHasPermissionOnItems(user *User, permissionKind, neededPermission string) *DB {
	itemsPerms := NewDataStore(newDB(conn.db.New())).PermissionsGenerated().
		WithPermissionForUser(user, permissionKind, neededPermission)
	return conn.Joins("JOIN ? AS rights ON rights.item_id = items.id", itemsPerms.SubQuery())
}

// WhereItemsAreVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsAreVisible(user *User) *DB {
	visibleItemsPerms := NewDataStore(newDB(conn.db.New())).PermissionsGenerated().VisibleToUser(user)
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
