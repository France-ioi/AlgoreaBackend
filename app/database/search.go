package database

import "strings"

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
func (conn *DB) WhereSearchStringMatches(field, searchString string) *DB {
	escapedSearchString := EscapeLikeString(searchString, '|')

	query := conn.db

	// For each word in the escaped search string.
	for _, word := range strings.Fields(escapedSearchString) {
		// Add a condition to the query to match the word.
		query = query.Where(field+" LIKE CONCAT('%', ?, '%') ESCAPE '|'", word)
	}

	return newDB(conn.ctx, query)
}
