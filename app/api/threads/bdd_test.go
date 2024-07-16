//go:build !unit

package threads_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func init() {
	testhelpers.BindGodogCmdFlags()
}

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}
