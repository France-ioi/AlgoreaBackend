package app_bdd_tests // nolint

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

// FeatureContext binds the supported steps to the verifying functions
func FeatureContext(s *godog.Suite) {
	ctx := &testContext{}
	s.BeforeScenario(ctx.setupTestContext)

	s.Step(`^the database has the following table \'([\w\-_]*)\':$`, ctx.dbHasTable)
	s.Step(`^a server is running as fallback$`, ctx.runFallbackServer)

	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, ctx.iSendrequestTo)
	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with the following body:$`, ctx.iSendrequestToWithBody)
	s.Step(`^the response code should be (\d+)$`, ctx.theResponseCodeShouldBe)
	s.Step(`^the response body should be, in JSON:$`, ctx.theResponseBodyShouldBeJSON)
	s.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, ctx.theResponseHeaderShouldBe)
	s.Step(`^it should be a JSON array with (\d+) entr(ies|y)$`, ctx.itShouldBeAJSONArrayWithEntries)
	s.Step(`^the table "([^"]*)" should be:$`, ctx.tableShouldBe)
}
