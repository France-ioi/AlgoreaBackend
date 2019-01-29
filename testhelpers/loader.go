package testhelpers

import (
	"flag"
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
}

// RunGodogTests launches GoDog tests (bdd tests) for the current directory
// (the one from the tested package)
func RunGodogTests(m *testing.M) {

	opt.Paths = []string{"."} // run feature files included in the package directory
	godog.BindFlags("godog.", flag.CommandLine, &opt)

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
