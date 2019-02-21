package answers_test

import (
	"fmt"
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestMain(m *testing.M) {

	if testhelpers.HasNoDBFlag() {
		fmt.Println("Skipping BDD tests in package 'groups' (no-db flag)")
		return
	}

	testhelpers.RunGodogTests(m)
}
