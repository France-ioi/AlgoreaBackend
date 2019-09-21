package database

// WhereUsersAreDescendantsOfGroup joins `groups_ancestors` on ancestor_group_id=groupID & child_group_id=users.self_group_id
func (conn *DB) WhereUsersAreDescendantsOfGroup(groupID int64) *DB {
	return conn.
		Joins("JOIN groups_ancestors ON groups_ancestors.child_group_id=users.self_group_id").
		Where("groups_ancestors.ancestor_group_id = ?", groupID)
}
