//go:build !prod

// Package testhelpers provides the interface and features to run the Gherkin tests.
package testhelpers

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
}

// BindGodogCmdFlags binds the command arguments into the Godog options.
func BindGodogCmdFlags() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

// RunGodogTests launches GoDog tests (bdd tests) for the current directory
// (the one from the tested package).
func RunGodogTests(t *testing.T, tags string) {
	appenv.SetDefaultEnvToTest()

	opt.Paths = featureFilesInCurrentDir()
	if tags != "" {
		opt.Tags = tags
	}

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opt,
	}

	assert.Zero(t, status)
}

func featureFilesInCurrentDir() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		panic(err)
	}
	var featuresFiles []string
	for _, f := range files {
		filename := f.Name()
		if filepath.Ext(filename) == ".feature" {
			featuresFiles = append(featuresFiles, filename)
		}
	}
	return featuresFiles
}
