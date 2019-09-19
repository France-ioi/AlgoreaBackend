package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors` on group_ancestor_id=groupID & group_child_id=users.group_self_id
func (conn *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return conn.
		Joins("JOIN groups_ancestors ON groups_ancestors.group_child_id=users.group_self_id").
		Where("groups_ancestors.group_ancestor_id = ?", groupID)
}
