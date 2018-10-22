package app_bdd_tests

import (
	"flag"
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
)

/*
This file allows the BDD (gherkin/cucumber) tests to be executed with `go test`.
`godog` is the name of the package we use to write Cucumber tests for Go.
You can use the options (flags) of the `godog` command by prefixing them with "godog.",
e.g., "go test -v --godog.format=progress --godog.random --godog.tags=wip"
*/

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {

	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
