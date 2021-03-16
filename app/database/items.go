package database

import "fmt"

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
	return conn.WhereGroupHasPermissionOnItems(user.GroupID, permissionKind, neededPermission)
}

// WhereItemsAreVisible returns a subview of the visible items for the given group basing on the given view
func (conn *DB) WhereItemsAreVisible(groupID int64) *DB {
	return conn.WhereGroupHasPermissionOnItems(groupID, "view", "info")
}

// WhereGroupHasPermissionOnItems returns a subview of the items
// on that the given group has `can_[permissionKind]_generated` >= `neededPermission`
// basing on the given view.
func (conn *DB) WhereGroupHasPermissionOnItems(groupID int64, permissionKind, neededPermission string) *DB {
	itemsPerms := NewDataStore(conn.New()).Permissions().
		MatchingGroupAncestors(groupID).
		Where("permissions.item_id = items.id").
		Where(fmt.Sprintf("%s >= ?", permissionColumnByKind(permissionKind)),
			NewDataStore(conn).PermissionsGranted().PermissionIndexByKindAndName(permissionKind, neededPermission)).
		Select("1").Limit(1)
	return conn.Where("EXISTS(?)", itemsPerms.QueryExpr())
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
