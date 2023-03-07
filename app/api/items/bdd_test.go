// +build !unit

package items_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}

func TestBDDWIP(t *testing.T) {
	testhelpers.RunGodogTests(t, "wop")
}
