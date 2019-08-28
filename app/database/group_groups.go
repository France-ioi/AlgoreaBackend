package database

// GroupRelationIsActiveCondition is an SQL condition restricting a value
// to be one of ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode')
const GroupRelationIsActiveCondition = " IN ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode')"

// WhereGroupRelationIsActive restricts `groups_groups.sType`
// to be one of ('direct', 'invitationAccepted', 'requestAccepted', 'joinedByCode')
func (conn *DB) WhereGroupRelationIsActive() *DB {
	return conn.Where("groups_groups.sType" + GroupRelationIsActiveCondition)
}
