// +build !unit

package auth_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestMain(m *testing.M) {
	app.RootDirectory = "../../../"
	testhelpers.RunGodogTests(m)
}
