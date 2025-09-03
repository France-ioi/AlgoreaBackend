//go:build !unit

package items_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}
