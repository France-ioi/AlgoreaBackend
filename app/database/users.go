package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors` on ancestor_group_id=groupID & child_group_id=users.group_id
func (conn *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return conn.
		Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id=users.group_id").
		Where("groups_ancestors_active.ancestor_group_id = ?", groupID)
}
