// +build !prod

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

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
}

// RunGodogTests launches GoDog tests (bdd tests) for the current directory
// (the one from the tested package)
func RunGodogTests(m *testing.M) {
	appenv.SetDefaultEnvToTest()

	opt.Paths = featureFilesInCurrentDir()
	godog.BindFlags("godog.", flag.CommandLine, &opt)

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
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
