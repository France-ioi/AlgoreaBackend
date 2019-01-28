package testhelpers

import (
	"os"
)

// HasNoDBFlag returns whether the test was called with env var TESTS_NODB set to a value.
// Use it to skip tests using the db
func HasNoDBFlag() bool {
	return len(os.Getenv("TESTS_NODB")) > 0
}
