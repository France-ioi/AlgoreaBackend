package database

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
func (conn *DB) WhereSearchStringMatches(field, searchString string) *DB {
	query := conn.db

	query = query.Where("MATCH ("+field+") AGAINST (? IN BOOLEAN MODE)", searchString)

	return newDB(conn.ctx, query)
}
