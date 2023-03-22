package database

// WhereGroupRelationIsActual restricts `groups_groups.type` to values corresponding to actual group membership and
// forces the relation to be not expired.
func (conn *DB) WhereGroupRelationIsActual() *DB {
	return conn.Where("NOW() < groups_groups.expires_at")
}
