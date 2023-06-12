package utils

import "unicode"

// Capitalize returns the string with its first character capitalized.
func Capitalize(str string) string {
	r := []rune(str)
	if len(r) < 1 {
		return ""
	}

	r[0] = unicode.ToUpper(r[0])

	return string(r)
}
