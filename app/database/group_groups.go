package database

// GroupRelationIsActiveCondition is an SQL condition restricting `groups_groups.sType` to values
// corresponding to actual group membership
const GroupRelationIsActiveCondition = " IN ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode')"

// WhereGroupRelationIsActive restricts `groups_groups.sType` to values corresponding to actual group membership
func (conn *DB) WhereGroupRelationIsActive() *DB {
	return conn.Where("groups_groups.sType" + GroupRelationIsActiveCondition)
}
