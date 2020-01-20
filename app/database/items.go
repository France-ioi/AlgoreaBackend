package database

// WhereUserHasViewPermissionOnItems returns a subview of the items
// on that the given user has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereUserHasViewPermissionOnItems(user *User, viewPermission string) *DB {
	return conn.WhereUserHasPermissionOnItems(user, "view", viewPermission)
}

// WhereGroupHasViewPermissionOnItems returns a subview of the items
// on that the given group has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereGroupHasViewPermissionOnItems(groupID int64, viewPermission string) *DB {
	return conn.WhereGroupHasPermissionOnItems(groupID, "view", viewPermission)
}

// WhereUserHasPermissionOnItems returns a subview of the items
// on that the given user has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereUserHasPermissionOnItems(user *User, permissionKind, neededPermission string) *DB {
	itemsPerms := NewDataStore(newDB(conn.db.New())).Permissions().
		WithPermissionForUser(user, permissionKind, neededPermission)
	return conn.Joins("JOIN ? AS permissions ON permissions.item_id = items.id", itemsPerms.SubQuery())
}

// WhereGroupHasPermissionOnItems returns a subview of the items
// on that the given group has `can_view_generated` >= `viewPermission`
// basing on the given view.
func (conn *DB) WhereGroupHasPermissionOnItems(groupID int64, permissionKind, neededPermission string) *DB {
	itemsPerms := NewDataStore(newDB(conn.db.New())).Permissions().
		WithPermissionForGroup(groupID, permissionKind, neededPermission)
	return conn.Joins("JOIN ? AS permissions ON permissions.item_id = items.id", itemsPerms.SubQuery())
}

// WhereItemsAreVisible returns a subview of the visible items for the given user basing on the given view
func (conn *DB) WhereItemsAreVisible(user *User) *DB {
	visibleItemsPerms := NewDataStore(newDB(conn.db.New())).Permissions().VisibleToUser(user)
	return conn.Joins("JOIN ? as visible ON visible.item_id = items.id", visibleItemsPerms.SubQuery())
}

// JoinsUserAndDefaultItemStrings joins items_strings with the given view twice
// (as default_strings for item's default language and as user_strings for the user's default language)
func (conn *DB) JoinsUserAndDefaultItemStrings(user *User) *DB {
	return conn.
		Joins(
			`LEFT JOIN items_strings default_strings
         ON default_strings.item_id = items.id AND default_strings.language_tag = items.default_language_tag`).
		Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.item_id=items.id AND user_strings.language_tag = ?`, user.DefaultLanguage)
}
