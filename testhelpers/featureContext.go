package testhelpers

import (
	"github.com/DATA-DOG/godog"
)

// FeatureContext binds the supported steps to the verifying functions
func FeatureContext(s *godog.Suite) {
	ctx := &TestContext{}
	s.BeforeScenario(ctx.SetupTestContext)

	s.Step(`^the database has the following table \'([\w\-_]*)\':$`, ctx.DBHasTable)
	s.Step(`^the table "([^"]*)" has a unique key "([^"]*)" on "([^"]*)"$`, ctx.TableHasUniqueKey)

	s.Step(`^a server is running as fallback$`, ctx.RunFallbackServer)
	s.Step(`^I am the user with ID "([^"]*)"$`, ctx.IAmUserWithID)
	s.Step(`^the time now is "([^"]*)"$`, ctx.TimeNow)
	s.Step(`^time is frozen$`, ctx.TimeIsFrozen)
	s.Step(`^the generated group password is "([^"]*)"$`, ctx.TheGeneratedGroupPasswordIs)
	s.Step(`^the generated group passwords are ("[^"]*"(?:\s*,\s*"[^"]*")*)$`, ctx.TheGeneratedGroupPasswordsAre)

	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, ctx.ISendrequestTo)
	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with the following body:$`, ctx.ISendrequestToWithBody)
	s.Step(`^the response code should be (\d+)$`, ctx.TheResponseCodeShouldBe)
	s.Step(`^the response body should be, in JSON:$`, ctx.TheResponseBodyShouldBeJSON)
	s.Step(`^the response body decoded as "([^"]+)" should be, in JSON:$`, ctx.TheResponseDecodedBodyShouldBeJSON)
	s.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, ctx.TheResponseHeaderShouldBe)
	s.Step(`^the response should be "([^"]*)"$`, ctx.TheResponseShouldBe)
	s.Step(`^the response error message should contain "(.*)"$`, ctx.TheResponseErrorMessageShouldContain)
	s.Step(`^it should be a JSON array with (\d+) entr(ies|y)$`, ctx.ItShouldBeAJSONArrayWithEntries)
	s.Step(`^the table "([^"]*)" should be:$`, ctx.TableShouldBe)
	s.Step(`^the table "([^"]*)" at IDs? "([^"]*)" should be:$`, ctx.TableAtIDShouldBe)
	s.Step(`^the table "([^"]*)" should not contain IDs? "([^"]*)"$`, ctx.TableShouldNotContainID)
	s.Step(`^the table "([^"]*)" should stay unchanged$`, ctx.TableShouldStayUnchanged)
	s.Step(`^the table "([^"]*)" should stay unchanged but the row with IDs? "([^"]*)"$`, ctx.TableShouldStayUnchangedButTheRowWithID)

	s.Step(`^logs should contain:$`, ctx.LogsShouldContain)

	s.AfterScenario(ctx.ScenarioTeardown)
}
