//go:build !unit

package app_test

import (
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
)

func init() {
	testhelpers.BindGodogCmdFlags()
}
