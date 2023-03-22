package database

// PermissionGeneratedStore implements database operations on `permissions_generated`.
type PermissionGeneratedStore struct {
	*DataStore
}

// MatchingUserAncestors returns a composable query of generated permissions matching groups of which the user is descendant.
func (s *PermissionGeneratedStore) MatchingUserAncestors(user *User) *DB {
	return s.MatchingGroupAncestors(user.GroupID)
}

// MatchingGroupAncestors returns a composable query of generated permissions matching groups of which the given group is descendant.
func (s *PermissionGeneratedStore) MatchingGroupAncestors(groupID int64) *DB {
	return s.Joins(`
		JOIN groups_ancestors_active AS ancestors
			ON ancestors.child_group_id = ? AND ancestors.ancestor_group_id = permissions.group_id`, groupID)
}

// AggregatedPermissionsForItemsOnWhichGroupHasViewPermission returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items on that the given group has `can_view_generated` >= `viewPermission`.
func (s *PermissionGeneratedStore) AggregatedPermissionsForItemsOnWhichGroupHasViewPermission(groupID int64, viewPermission string) *DB {
	return s.AggregatedPermissionsForItemsOnWhichGroupHasPermission(groupID, "view", viewPermission)
}

// AggregatedPermissionsForItemsOnWhichGroupHasPermission returns a composable query for getting access rights
// (as *_generated_value) and item ids (as item_id)
// for all the items on that the given group has 'permissionKind' >= `neededPermission`.
func (s *PermissionGeneratedStore) AggregatedPermissionsForItemsOnWhichGroupHasPermission(
	groupID int64, permissionKind, neededPermission string) *DB {
	return s.AggregatedPermissionsForItems(groupID).
		HavingMaxPermissionAtLeast(permissionKind, neededPermission)
}

// AggregatedPermissionsForItems returns a composable query for getting access rights of the given group
// (as *_generated_value) and item ids (as item_id) for all items.
func (s *PermissionGeneratedStore) AggregatedPermissionsForItems(groupID int64) *DB {
	return s.
		Select(`
			permissions.item_id,
			MAX(can_view_generated_value) AS can_view_generated_value,
			MAX(can_grant_view_generated_value) AS can_grant_view_generated_value,
			MAX(can_watch_generated_value) AS can_watch_generated_value,
			MAX(can_edit_generated_value) AS can_edit_generated_value,
			MAX(is_owner_generated) AS is_owner_generated`).
		Joins("JOIN groups_ancestors_active AS ancestors ON ancestors.ancestor_group_id = permissions.group_id").
		Where("ancestors.child_group_id = ?", groupID).
		Group("permissions.item_id")
}

// AggregatedPermissionsForItemsVisibleToGroup returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items that are visible to the given group.
func (s *PermissionGeneratedStore) AggregatedPermissionsForItemsVisibleToGroup(groupID int64) *DB {
	return s.AggregatedPermissionsForItemsOnWhichGroupHasViewPermission(groupID, "info")
}
