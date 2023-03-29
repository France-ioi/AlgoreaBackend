//go:build !unit

package app_test

import (
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func init() {
	testhelpers.BindGodogCmdFlags()
}
