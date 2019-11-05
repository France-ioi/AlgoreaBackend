package database

// PermissionGeneratedStore implements database operations on `permissions_generated`
type PermissionGeneratedStore struct {
	*DataStore
}

// MatchingUserAncestors returns a composable query of generated permissions matching groups of which the user is member
func (s *PermissionGeneratedStore) MatchingUserAncestors(user *User) *DB {
	db := s.GroupAncestors().UserAncestors(user)
	userAncestors := db.SubQuery()
	return s.Joins("JOIN ? AS ancestors ON permissions_generated.group_id = ancestors.ancestor_group_id", userAncestors)
}

// AccessRightsForItemsVisibleToGroup returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items that are visible to the given group.
// Note that the `groupID` can be nil.
func (s *PermissionGeneratedStore) AccessRightsForItemsVisibleToGroup(groupID *int64) *DB {
	return s.
		Select(`
			item_id,
			MAX(can_view_generated_value) AS can_view_generated_value`).
		Joins(`
			JOIN (
				SELECT * FROM groups_ancestors_active
				WHERE groups_ancestors_active.child_group_id = ?
			) AS ancestors
			ON ancestors.ancestor_group_id = permissions_generated.group_id`, groupID).
		Group("permissions_generated.item_id").
		Having("can_view_generated_value > ?", s.PermissionsGranted().ViewIndexByKind("none"))
}

// AccessRightsForItemsVisibleToUser returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items that are visible to the given user.
func (s *PermissionGeneratedStore) AccessRightsForItemsVisibleToUser(user *User) *DB {
	return s.AccessRightsForItemsVisibleToGroup(&user.GroupID)
}
