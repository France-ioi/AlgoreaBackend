package database

import "strings"

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
// If fallbackField is not empty, it is used as a fallback if the field is NULL.
func (conn *DB) WhereSearchStringMatches(field, fallbackField, searchString string) *DB {
	query := conn.db

	words := strings.Fields(strings.Trim(searchString, " "))

	for i := 0; i < len(words); i++ {
		word := words[i]
		if word == "" {
			continue
		}

		// The "+" sign means that the word must be present in the result.
		if word[0] != '+' {
			word = "+" + word
		}

		// The "*" sign means that the word can be a prefix of a word in the result.
		if word[len(word)-1] != '*' {
			word += "*"
		}

		words[i] = word
	}

	searchPattern := strings.Join(words, " ")

	condition := "(" + field + " IS NOT NULL AND MATCH(" + field + ") AGAINST(? IN BOOLEAN MODE))"
	if fallbackField == "" {
		query = query.Where(condition, searchPattern)
	} else {
		condition += "OR (" + field + " IS NULL AND MATCH(" + fallbackField + ") AGAINST(? IN BOOLEAN MODE))"
		query = query.Where(condition, searchPattern, searchPattern)
	}

	return newDB(conn.ctx, query)
}
