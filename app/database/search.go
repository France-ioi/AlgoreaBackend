package database

import "strings"

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
func (conn *DB) WhereSearchStringMatches(field, searchString string) *DB {
	query := conn.db

	// Remove all the special characters from the search string.
	searchString = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return ' '
	}, searchString)

	// Remove consecutive spaces.
	searchString = strings.Join(strings.Fields(searchString), " ")

	words := strings.Fields(strings.Trim(searchString, " "))

	// Search words than begin with.
	for i := 0; i < len(words); i++ {
		word := words[i]

		// The "+" sign means that the word must be present in the result.
		// The "*" sign means that the word can be a prefix of a word in the result.
		words[i] = "+" + word + "*"
	}

	query = query.Where("MATCH ("+field+") AGAINST (? IN BOOLEAN MODE)", strings.Join(words, " "))

	return newDB(conn.ctx, query)
}
