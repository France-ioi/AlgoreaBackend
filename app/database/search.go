package database

import (
	"strings"
	"unicode"
)

// WhereSearchStringMatches returns a composable query where {field} matches the search string {searchString}
// All the words in the search string are matched with "AND".
// If fallbackField is not empty, it is used as a fallback if the field is NULL.
//
// We use the MySQL fulltext search with innodb_ft_min_token_size=1 and an empty stopword list.
// This method would have to filter out short words and stopwords from the search string if we used different settings.
func (conn *DB) WhereSearchStringMatches(field, fallbackField, searchString string) *DB {
	query := conn.db

	// Remove all the special characters from the search string.
	searchString = strings.Map(func(r rune) rune {
		// Keep only letters (for all the world languages), digits, and underscores.
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}

		return ' '
	}, searchString)

	words := strings.Fields(strings.TrimSpace(searchString))
	if len(words) == 0 {
		// If the search string has no words, we return an empty result.
		return conn.Where("FALSE")
	}

	for wordIndex := 0; wordIndex < len(words); wordIndex++ {
		word := words[wordIndex]

		// The "+" sign means that the word must be present in the result.
		word = "+" + word

		// The "*" sign means that the word can be a prefix of a word in the result.
		word += "*"

		words[wordIndex] = word
	}

	searchPattern := strings.Join(words, " ")

	condition := "(" + field + " IS NOT NULL AND MATCH(" + field + ") AGAINST(? IN BOOLEAN MODE))"
	if fallbackField == "" {
		query = query.Where(condition, searchPattern)
	} else {
		condition += " OR (" + field + " IS NULL AND MATCH(" + fallbackField + ") AGAINST(? IN BOOLEAN MODE))"
		query = query.Where(condition, searchPattern, searchPattern)
	}

	return newDB(query, conn.ctes)
}
