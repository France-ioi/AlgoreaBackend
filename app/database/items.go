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

// WhereUserHaveStartedResultOnItem makes sure that the user have a started result on the item, whatever the attempt.
func (conn *DB) WhereUserHaveStartedResultOnItem(user *User) *DB {
	return conn.
		Joins(`
				JOIN results AS current_user_results
					ON current_user_results.item_id = items.id AND
					current_user_results.participant_id = ?`, user.GroupID).
		Where("current_user_results.started = 1")
}

// WhereItemsAreVisible returns a subview of the visible items for the given group basing on the given view.
func (conn *DB) WhereItemsAreVisible(groupID int64) *DB {
	return conn.WhereGroupHasPermissionOnItems(groupID, "view", "info")
}

// WhereItemsContentAreVisible returns a subview of the content visible items for the given group.
func (conn *DB) WhereItemsContentAreVisible(groupID int64) *DB {
	return conn.WhereGroupHasPermissionOnItems(groupID, "view", "content")
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
// (as default_strings for item's default language and as user_strings for the user's default language).
func (conn *DB) JoinsUserAndDefaultItemStrings(user *User) *DB {
	return conn.
		Joins(
			`LEFT JOIN items_strings default_strings
         ON default_strings.item_id = items.id AND default_strings.language_tag = items.default_language_tag`).
		Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.item_id=items.id AND user_strings.language_tag = ?`, user.DefaultLanguage)
}
