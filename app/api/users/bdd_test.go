//go:build !unit

package users_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func init() {
	testhelpers.BindGodogCmdFlags()
}

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}
