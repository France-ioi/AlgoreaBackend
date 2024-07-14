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

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
)

var defaultGodogOptions = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
}

var godogFlagsBound bool

// BindGodogCmdFlags binds the command arguments into the Godog options.
func BindGodogCmdFlags() {
	if godogFlagsBound {
		return
	}
	godog.BindFlags("godog.", flag.CommandLine, &defaultGodogOptions)
	godogFlagsBound = true
}

// RunGodogTests launches GoDog tests (bdd tests) for the current directory
// (the one from the tested package).
func RunGodogTests(t *testing.T, tags string) {
	appenv.SetDefaultEnvToTest()

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	options := defaultGodogOptions
	featureFiles := featureFilesInCurrentDir()

	if tags != "" {
		options.Tags = tags
	}

	for _, featureFile := range featureFiles {
		featureFile := featureFile
		t.Run(featureFile, func(t *testing.T) {
			featureFileOptions := options
			featureFileOptions.Paths = []string{featureFile}
			featureFileOptions.TestingT = t // Testing instance that will run subtests.
			suite := godog.TestSuite{
				ScenarioInitializer: InitializeScenario,
				Options:             &featureFileOptions,
			}
			suite.Run()
		})
	}
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
