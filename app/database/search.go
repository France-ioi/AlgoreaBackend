package database

import (
	"strings"
	"unicode"
)

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
// If fallbackField is not empty, it is used as a fallback if the field is NULL.
func (conn *DB) WhereSearchStringMatches(field, fallbackField, searchString string) *DB {
	query := conn.db

	// Remove all the special characters from the search string.
	searchString = strings.Map(func(r rune) rune {
		// Keep only letters (for all the world languages), digits, and apostrophes.
		if r == '\'' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}

		return ' '
	}, searchString)

	words := strings.Fields(strings.Trim(searchString, " "))

	for i := 0; i < len(words); i++ {
		word := words[i]

		// The "+" sign means that the word must be present in the result.
		word = "+" + word

		// The "*" sign means that the word can be a prefix of a word in the result.
		word += "*"

		words[i] = word
	}

	searchPattern := strings.Join(words, " ")

	condition := "(" + field + " IS NOT NULL AND MATCH(" + field + ") AGAINST(? IN BOOLEAN MODE))"
	if fallbackField == "" {
		query = query.Where(condition, searchPattern)
	} else {
		condition += " OR (" + field + " IS NULL AND MATCH(" + fallbackField + ") AGAINST(? IN BOOLEAN MODE))"
		query = query.Where(condition, searchPattern, searchPattern)
	}

	return newDB(conn.ctx, query, conn.ctes, conn.logConfig)
}
