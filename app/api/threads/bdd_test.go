//go:build !unit

package threads_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}
