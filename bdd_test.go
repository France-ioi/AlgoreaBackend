package main

import (
	"github.com/DATA-DOG/godog"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

// FeatureContext is required to be able to use the `godog` command at root of the project
//noinspection GoUnusedExportedFunction
func FeatureContext(s *godog.Suite) { // nolint: megacheck, deadcode
	testhelpers.FeatureContext(s)
}
