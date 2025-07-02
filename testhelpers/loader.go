//go:build !prod

// Package testhelpers provides the interface and features to run the Gherkin tests.
package testhelpers

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/zenovich/flowmingo"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
)

var defaultGodogOptions = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
	Strict: true,
}

func init() { //nolint:gochecknoinits // bind Godog flags only once before running tests
	bindGodogCmdFlags()
}

// bindGodogCmdFlags binds the command arguments into the Godog options.
func bindGodogCmdFlags() {
	godog.BindFlags("godog.", flag.CommandLine, &defaultGodogOptions)
}

type contextKey string

var outputRestorerFuncKey = contextKey("outputRestorerFunc")

// RunGodogTests launches GoDog tests (bdd tests) for the current directory
// (the one from the tested package).
func RunGodogTests(t *testing.T, tags string) {
	t.Helper()

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
			var restoreFunc flowmingo.RestoreFunc
			if !testing.Verbose() { // Do not suppress output in verbose mode
				restoreFunc = flowmingo.CaptureStdoutAndStderr() // Suppress the output of the suite
				t.Cleanup(func() {
					if restoreFunc != nil {
						restoreFunc(t.Failed()) // Do not pass through the output of a passed suite
					}
				})
			}

			featureFileOptions := options
			featureFileOptions.Paths = []string{featureFile}
			featureFileOptions.TestingT = t // Testing instance that will run subtests.
			// Set the context with the output restorer function to allow children to restore the output immediately when test fails.
			featureFileOptions.DefaultContext = context.WithValue(context.Background(), outputRestorerFuncKey, &restoreFunc)
			suite := godog.TestSuite{
				ScenarioInitializer: InitializeScenario,
				Options:             &featureFileOptions,
			}
			suite.Run()
		})
	}
}

func featureFilesInCurrentDir() []string {
	files, err := os.ReadDir(".")
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
