package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors` on idGroupAncestor=groupID & idGroupChild=users.idGroupSelf
func (s *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return s.
		Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild=users.idGroupSelf").
		Where("groups_ancestors.idGroupAncestor = ?", groupID)
}
