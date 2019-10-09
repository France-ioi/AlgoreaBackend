package database

// GroupRelationIsActiveCondition is an SQL condition restricting `groups_groups.type` to values
// corresponding to actual group membership
const GroupRelationIsActiveCondition = " IN ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode')"

// WhereGroupRelationIsActual restricts `groups_groups.type` to values corresponding to actual group membership and
// forces the relation to be not expired
func (conn *DB) WhereGroupRelationIsActual() *DB {
	return conn.Where("groups_groups.type" + GroupRelationIsActiveCondition + " AND NOW() < groups_groups.expires_at")
}

// WhereActiveGroupRelationIsActual restricts `groups_groups_active.type` to values corresponding to actual group membership
func (conn *DB) WhereActiveGroupRelationIsActual() *DB {
	return conn.Where("groups_groups_active.type" + GroupRelationIsActiveCondition)
}
