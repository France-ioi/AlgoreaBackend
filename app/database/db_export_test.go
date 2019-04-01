package database

func (conn *DB) Exec(sql string, values ...interface{}) *DB {
	return newDB(conn.db.Exec(sql, values...))
}
