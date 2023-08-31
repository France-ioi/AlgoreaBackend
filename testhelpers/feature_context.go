//go:build !prod

package testhelpers

import (
	"github.com/cucumber/godog"
)

// FeatureContext binds the supported steps to the verifying functions.
func FeatureContext(s *godog.Suite) {
	ctx := &TestContext{}
	s.BeforeScenario(ctx.SetupTestContext)

	s.Step(`^the template constant "([^"]+)" is "(.*)"$`, ctx.TheTemplateConstantIsString)
	s.Step(`^the template constant "([^"]+)" is:$`, ctx.TheTemplateConstantIsDocString)

	s.Step(`^the database has the following table \'([\w\-_]*)\':$`, ctx.DBHasTable)
	s.Step(`^the database table \'([\w\-_]*)\' has also the following rows?:$`, ctx.DBHasTable)
	s.Step(`^the database has the following users:$`, ctx.DBHasUsers)
	s.Step(`^the groups ancestors are computed$`, ctx.DBGroupsAncestorsAreComputed)

	s.Step(`^I am (@\w+)$`, ctx.IAm)
	s.Step(`^I am the user with id "([^"]*)"$`, ctx.IAmUserWithID)
	s.Step(`^there is a user (@\w+)$`, ctx.ThereIsAUser)
	s.Step(`^there are the following users:$`, ctx.ThereAreTheFollowingUsers)
	s.Step(`^the time now is "([^"]*)"$`, ctx.TimeNow)
	s.Step(`^time is frozen$`, ctx.TimeIsFrozen)
	s.Step(`^the generated group code is "([^"]*)"$`, ctx.TheGeneratedGroupCodeIs)
	s.Step(`^the generated group codes are ("[^"]*"(?:\s*,\s*"[^"]*")*)$`, ctx.TheGeneratedGroupCodesAre)
	s.Step(`^the generated auth key is "([^"]*)"$`, ctx.TheGeneratedAuthKeyIs)
	s.Step(`^the application config is:$`, ctx.TheApplicationConfigIs)
	s.Step(`^the context variable "([^"]*)" is "([^"]*)"$`, ctx.TheContextVariableIs)

	s.Step(`^there are the following groups:$`, ctx.ThereAreTheFollowingGroups)
	s.Step(`^there is a group with "([^"]*)"$`, ctx.ThereIsAGroupWith)
	s.Step(`^there is a group (@\w+)$`, ctx.ThereIsAGroup)
	s.Step(`^I am a member of the group (@\w+)$`, ctx.IAmAMemberOfTheGroup)
	s.Step(`^I am a member of the group with id "([^"]*)"$`, ctx.IAmAMemberOfTheGroupWithID)
	s.Step(`^(@\w+) is a member of the group (@\w+)$`, ctx.UserIsAMemberOfTheGroup)
	s.Step(
		`^(@\w+) is a member of the group (@\w+) who has approved access to his personal info$`,
		ctx.UserIsAMemberOfTheGroupWhoHasApprovedAccessToHisPersonalInfo,
	)
	s.Step(`allUsersGroup is defined as the group (@\w+)$`, ctx.AllUsersGroupIsDefinedAsTheGroup)

	s.Step(`^I am a manager of the group with id "([^"]*)"$`, ctx.IAmAManagerOfTheGroupWithID)
	s.Step(`^I am a manager of the group (@\w+)$`, ctx.IAmAManagerOfTheGroup)
	s.Step(`^(@\w+) is a manager of the group (@\w+) and can watch its members$`, ctx.UserIsAManagerOfTheGroupAndCanWatchItsMembers)
	s.Step(`^I am a manager of the group (@\w+) and can watch its members$`, ctx.IAmAManagerOfTheGroupAndCanWatchItsMembers)
	s.Step(`(@\w+) is a manager of the group (@\w+) and can grant group access`, ctx.UserIsAManagerOfTheGroupAndCanGrantGroupAccess)
	s.Step(`^the group (@\w+) is a descendant of the group (@\w+)$`, ctx.theGroupIsADescendantOfTheGroup)
	s.Step(`^there are the following items:$`, ctx.ThereAreTheFollowingItems)
	s.Step(`^there are the following tasks:$`, ctx.ThereAreTheFollowingTasks)
	s.Step(`^there are the following item permissions:$`, ctx.ThereAreTheFollowingItemPermissions)
	s.Step(`^there are the following item relations:$`, ctx.ThereAreTheFollowingItemRelations)
	s.Step(`^I can watch the group (@\w+)$`, ctx.ICanWatchGroup)
	s.Step(`^I can watch the participant with id "([^"]*)"$`, ctx.ICanWatchGroupWithID)
	s.Step(`^I can view (none|info|content|content_with_descendants|solution) on item with id "([^"]*)"$`,
		ctx.ICanViewOnItemWithID)
	s.Step(`^I can watch (none|result|answer|answer_with_grant) on item with id "([^"]*)"$`, ctx.ICanWatchOnItemWithID)
	s.Step(`^I can request help to the group with id "([^"]*)" on the item with id "([^"]*)"$`,
		ctx.ICanRequestHelpToTheGroupWithIDOnTheItemWithID)

	s.Step(`^there are the following results:$`, ctx.ThereAreTheFollowingResults)
	s.Step(`^I have validated the item with id "([^"]*)"$`, ctx.IHaveValidatedItemWithID)

	s.Step(`^there are the following threads:$`, ctx.ThereAreTheFollowingThreads)
	s.Step(`^there is a thread with "([^"]*)"$`, ctx.ThereIsAThreadWith)
	s.Step(`^there is no thread with "([^"]*)"$`, ctx.ThereIsNoThreadWith)
	s.Step(`^I am part of the helper group of the thread$`, ctx.IAmPartOfTheHelperGroupOfTheThread)

	s.Step(`^the "([^"]*)" request header is "(.*)"$`, ctx.TheRequestHeaderIs)
	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, ctx.ISendrequestTo)
	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with the following body:$`, ctx.ISendrequestToWithBody)
	s.Step(`^the response code should be (\d+)$`, ctx.TheResponseCodeShouldBe)
	s.Step(`^the response body should be, in JSON:$`, ctx.TheResponseBodyShouldBeJSON)
	s.Step(`^the response body should be:$`, ctx.TheResponseBodyShouldBe)
	s.Step(`^the response body decoded as "([^"]+)" should be, in JSON:$`, ctx.TheResponseDecodedBodyShouldBeJSON)
	s.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, ctx.TheResponseHeaderShouldBe)
	s.Step(`^the response headers? "([^"]*)" should be:`, ctx.TheResponseHeadersShouldBe)
	s.Step(`^the response should be "([^"]*)"$`, ctx.TheResponseShouldBe)
	s.Step(`^the response error message should contain "(.*)"$`, ctx.TheResponseErrorMessageShouldContain)

	s.Step(`^the response should be a JSON array with (\d+) entr(ies|y)$`, ctx.ItShouldBeAJSONArrayWithEntries)
	s.Step(`^the response should be a ZIP file containing the following files:$`, ctx.ItShouldBeAZIPFileContainingTheFollowingFiles)
	s.Step(`^the response at ([^ ]+) should be "([^"]*)"$`, ctx.TheResponseAtShouldBeTheValue)
	s.Step("^the response at ([^ ]+) should be:$", ctx.TheResponseAtShouldBe)
	s.Step("^the response at ([^ ]+) in JSON should be:$", ctx.TheResponseAtInJSONShouldBe)

	s.Step(`^the table "([^"]*)" should be:$`, ctx.TableShouldBe)
	s.Step(`^the table "([^"]*)" should be empty$`, ctx.TableShouldBeEmpty)
	s.Step(`^the table "([^"]*)" at (\w+?)s? "([^"]*)" should be:$`, ctx.TableAtColumnValueShouldBe)
	s.Step(`^the table "([^"]*)" at (\w+?)s? "([^"]*)" should be empty$`, ctx.TableAtColumnValueShouldBeEmpty)
	s.Step(`^the table "([^"]*)" should not contain (\w+?)s? "([^"]*)"$`, ctx.TableShouldNotContainColumnValue)
	s.Step(`^the table "([^"]*)" should stay unchanged$`, ctx.TableShouldStayUnchanged)
	s.Step(`^the table "([^"]*)" should stay unchanged but the rows? with (\w+?)s? "([^"]*)"$`,
		ctx.TableShouldStayUnchangedButTheRowWithColumnValue)
	s.Step(`^the table "([^"]*)" should stay unchanged but the rows? with ([^"]*) "([^"]*)" should be deleted$`,
		ctx.TableShouldStayUnchangedButTheRowsWithColumnValueShouldBeDeleted)
	s.Step(`^the DB time now is "([^"]*)"$`, ctx.DbTimeNow)

	s.Step(`^"([^"]+)" is a token signed by (.+) with the following payload:$`, ctx.SignedTokenIsDistributed)
	s.Step(`^logs should contain:$`, ctx.LogsShouldContain)

	s.Step(`^the login module "token" endpoint for code "([^"]*)" returns (\d+) with body:$`, ctx.TheLoginModuleTokenEndpointForCodeReturns)
	s.Step(`^the login module "token" endpoint for code "([^"]*)" and code_verifier "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForCodeAndCodeVerifierReturns)
	s.Step(
		`^the login module "token" endpoint for code "([^"]*)" and code_verifier "([^"]*)" and redirect_uri "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForCodeAndCodeVerifierAndRedirectURIReturns)
	s.Step(`^the login module "token" endpoint for refresh token "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForRefreshTokenReturns)
	s.Step(`^the login module "account" endpoint for token "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleAccountEndpointForTokenReturns)
	s.Step(`^the login module "unlink_client" endpoint for user id "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleUnlinkClientEndpointForUserIDReturns)
	s.Step(`^the login module "create" endpoint with params "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleCreateEndpointWithParamsReturns)
	s.Step(`^the login module "delete" endpoint with params "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleDeleteEndpointWithParamsReturns)
	s.Step(
		`^the login module "lti_result/send" endpoint for user id "([^"]*)", `+
			`content id "([^"]*)", score "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleLTIResultSendEndpointForUserIDContentIDScoreReturns)

	s.AfterScenario(ctx.ScenarioTeardown)
}
