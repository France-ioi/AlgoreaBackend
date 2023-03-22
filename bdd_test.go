// +build !unit

package main_test

import (
	"github.com/cucumber/godog"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

// FeatureContext is required to be able to use the `godog` command at root of the project
// noinspection GoUnusedExportedFunction.
func FeatureContext(s *godog.Suite) { // nolint
	testhelpers.FeatureContext(s)
}
