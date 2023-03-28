//go:build !unit

package api_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func init() {
	testhelpers.InitGodogCommandLineFlags()
}

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t, "")
}
