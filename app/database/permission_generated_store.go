package database

// PermissionGeneratedStore implements database operations on `permissions_generated`
type PermissionGeneratedStore struct {
	*DataStore
}

// MatchingUserAncestors returns a composable query of generated permissions matching groups of which the user is member
func (s *PermissionGeneratedStore) MatchingUserAncestors(user *User) *DB {
	db := s.GroupAncestors().UserAncestors(user)
	userAncestors := db.SubQuery()
	return s.Joins("JOIN ? AS ancestors ON permissions.group_id = ancestors.ancestor_group_id", userAncestors)
}

// WithViewPermissionForGroup returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items on that the given group has `can_view_generated` >= `viewPermission`.
// Note that the `groupID` can be nil.
func (s *PermissionGeneratedStore) WithViewPermissionForGroup(groupID int64, viewPermission string) *DB {
	return s.WithPermissionForGroup(groupID, "view", viewPermission)
}

// WithPermissionForGroup returns a composable query for getting access rights
// (as *_generated_value) and item ids (as item_id)
// for all the items on that the given group has 'permissionKind' >= `neededPermission`.
// Note that the `groupID` can be nil.
func (s *PermissionGeneratedStore) WithPermissionForGroup(groupID int64, permissionKind, neededPermission string) *DB {
	return s.
		Select(`
			item_id,
			MAX(can_view_generated_value) AS can_view_generated_value`).
		Joins(`
			JOIN (
				SELECT * FROM groups_ancestors_active
				WHERE groups_ancestors_active.child_group_id = ?
			) AS ancestors
			ON ancestors.ancestor_group_id = permissions.group_id`, groupID).
		WherePermissionIsAtLeast(permissionKind, neededPermission).
		Group("permissions.item_id")
}

// VisibleToGroup returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items that are visible to the given group.
// Note that the `groupID` can be nil.
func (s *PermissionGeneratedStore) VisibleToGroup(groupID int64) *DB {
	return s.WithViewPermissionForGroup(groupID, "info")
}

// VisibleToUser returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items that are visible to the given user.
func (s *PermissionGeneratedStore) VisibleToUser(user *User) *DB {
	return s.VisibleToGroup(user.GroupID)
}

// WithViewPermissionForUser returns a composable query for getting access rights
// (as can_view_generated_value) and item ids (as item_id)
// for all the items on that the given user has `can_view_generated` >= `viewPermission`.
func (s *PermissionGeneratedStore) WithViewPermissionForUser(user *User, viewPermission string) *DB {
	return s.WithViewPermissionForGroup(user.GroupID, viewPermission)
}

// WithPermissionForUser returns a composable query for getting access rights
// (as *_generated_value) and item ids (as item_id)
// for all the items on that the given user has 'permissionKind' >= `neededPermission`.
func (s *PermissionGeneratedStore) WithPermissionForUser(user *User, permissionKind, neededPermission string) *DB {
	return s.WithPermissionForGroup(user.GroupID, permissionKind, neededPermission)
}
