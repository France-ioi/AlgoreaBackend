package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors` on idGroupAncestor=groupID & idGroupChild=users.idGroupSelf
func (conn *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return conn.
		Joins("JOIN groups_ancestors ON groups_ancestors.idGroupChild=users.idGroupSelf").
		Where("groups_ancestors.idGroupAncestor = ?", groupID)
}
