package api_test

import (
	"fmt"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestMain(m *testing.M) {

	if testhelpers.HasNoDBFlag() {
		fmt.Println("Skipping BDD tests in package 'api' (TESTS_NODB env set)")
		return
	}

	testhelpers.RunGodogTests(m)
}
