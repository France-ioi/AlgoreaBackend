// +build !unit

package groups_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestMain(m *testing.M) {
	app.RootDirectory = "../../../" // nolint:goconst
	testhelpers.RunGodogTests(m)
}
