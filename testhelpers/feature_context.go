//go:build !prod && !unit

package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/zenovich/flowmingo"
)

// InitializeScenario binds the supported steps to the verifying functions.
func InitializeScenario(scenarioContext *godog.ScenarioContext) {
	var restoreFunc flowmingo.RestoreFunc
	ctx := &TestContext{}

	scenarioContext.Before(func(contextCtx context.Context, sc *godog.Scenario) (context.Context, error) {
		if !testing.Verbose() { // Do not suppress output in verbose mode
			restoreFunc = flowmingo.CaptureStdoutAndStderr() // Suppress the output of the scenario
		}
		ctx.SetupTestContext(sc)
		return contextCtx, nil
	})

	scenarioContext.Step(`^the template constant "([^"]+)" is "(.*)"$`, ctx.TheTemplateConstantIsString)
	scenarioContext.Step(`^the template constant "([^"]+)" is:$`, ctx.TheTemplateConstantIsDocString)

	scenarioContext.Step(`^the database has the following table "([^"]+)":$`, ctx.DBHasTable)
	scenarioContext.Step(`^the database table "([^"]+)"(?: also)? has the following rows?:$`, ctx.DBHasTable)
	scenarioContext.Step(`^the database(?: also)? has the following users?:$`, ctx.DBHasUsers)
	scenarioContext.Step(`^the groups ancestors are computed$`, ctx.DBGroupsAncestorsAreComputed)
	scenarioContext.Step(`^the generated permissions are computed$`, ctx.DBGeneratedPermissionsAreComputed)
	scenarioContext.Step(`^the results are computed$`, ctx.DBResultsAreComputed)

	ctx.registerFeaturesForSessions(scenarioContext)
	ctx.registerFeaturesForUsers(scenarioContext)

	scenarioContext.Step(`^the time now is "([^"]*)"$`, ctx.TimeNow)
	scenarioContext.Step(`^the server time now is "([^"]*)"$`, ctx.ServerTimeNow)
	scenarioContext.Step(`^the time is frozen$`, ctx.TimeIsFrozen)
	scenarioContext.Step(`^the server time is frozen$`, ctx.ServerTimeIsFrozen)
	scenarioContext.Step(`^the generated group code is "([^"]*)"$`, ctx.TheGeneratedGroupCodeIs)
	scenarioContext.Step(`^the generated group codes are ("[^"]*"(?:\s*,\s*"[^"]*")*)$`, ctx.TheGeneratedGroupCodesAre)
	scenarioContext.Step(`^the generated auth key is "([^"]*)"$`, ctx.TheGeneratedAuthKeyIs)
	scenarioContext.Step(`^the application config is:$`, ctx.TheApplicationConfigIs)
	scenarioContext.Step(`^the context variable "([^"]*)" is "([^"]*)"$`, ctx.TheContextVariableIs)

	ctx.registerFeaturesForGroups(scenarioContext)

	scenarioContext.Step(`^I am a manager of the group with id "([^"]*)"$`, ctx.IAmAManagerOfTheGroupWithID)
	scenarioContext.Step(`^I am a manager of the group (@\w+)$`, ctx.IAmAManagerOfTheGroup)
	scenarioContext.Step(`^I am a manager of the group (.+) and can watch for submissions from the group and its descendants$`,
		ctx.IAmAManagerOfTheGroupAndCanWatchItsMembers)
	scenarioContext.Step(
		`^the group (@\w+) is a manager of the group (@\w+) and can watch for submissions from the group and its descendants$`,
		ctx.GroupIsAManagerOfTheGroupAndCanWatchItsMembers)
	scenarioContext.Step(`^the group (@\w+) is a manager of the group (@\w+) and can grant group access`,
		ctx.GroupIsAManagerOfTheGroupAndCanGrantGroupAccess)
	scenarioContext.Step(
		`^the group (@\w+) is a manager of the group (@\w+) and can manage memberships and the group`,
		ctx.GroupIsAManagerOfTheGroupAndCanManageMembershipsAndGroup,
	)

	scenarioContext.Step(`^there are the following items:$`, ctx.ThereAreTheFollowingItems)
	scenarioContext.Step(`^there are the following tasks:$`, ctx.ThereAreTheFollowingTasks)
	scenarioContext.Step(`^there are the following item permissions:$`, ctx.ThereAreTheFollowingItemPermissions)
	scenarioContext.Step(`^there are the following item relations:$`, ctx.ThereAreTheFollowingItemRelations)
	scenarioContext.Step(`^I can view (none|info|content|content_with_descendants|solution) of the item (.+)$`,
		ctx.IHaveViewPermissionOnItem)
	scenarioContext.Step(`^the group (\@\w+) can view (none|info|content|content_with_descendants|solution) of the item (.+)$`,
		ctx.GroupHasViewPermissionOnItem)
	scenarioContext.Step(`^I have the watch permission set to "(none|result|answer|answer_with_grant)" on the item (.+)$`,
		ctx.IHaveWatchPermissionOnItem)
	scenarioContext.Step(`^the group (.+) has the watch permission set to "(none|result|answer|answer_with_grant)" on the item (.+)$`,
		ctx.GroupHasWatchPermissionOnItem)
	scenarioContext.Step(`^I can request help to the group with id "([^"]*)" on the item with id "([^"]*)"$`,
		ctx.ICanRequestHelpToTheGroupWithIDOnTheItemWithID)

	ctx.registerFeaturesForGroupMembershipChanges(scenarioContext)
	ctx.registerFeaturesForGroupPendingRequests(scenarioContext)

	scenarioContext.Step(`^there are the following validated results:$`, ctx.ThereAreTheFollowingValidatedResults)
	scenarioContext.Step(`^I have a validated result on the item (.+)$`, ctx.IHaveValidatedResultOnItem)
	scenarioContext.Step(`^the group (.+) has a validated result on the item (.+)$`, ctx.GroupHasValidatedResultOnItem)

	scenarioContext.Step(`^there are the following threads:$`, ctx.ThereAreTheFollowingThreads)
	scenarioContext.Step(`^there is a thread with "(.*)"$`, ctx.ThereIsAThreadWith)
	scenarioContext.Step(`^there is no thread with "([^"]*)"$`, ctx.ThereIsNoThreadWith)
	scenarioContext.Step(`^I am part of the helper group of the thread$`, ctx.IAmPartOfTheHelperGroupOfTheThread)

	scenarioContext.Step(`^the "([^"]*)" request header is "(.*)"$`, ctx.TheRequestHeaderIs)
	scenarioContext.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, ctx.ISendrequestTo)
	scenarioContext.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with the following body:$`, ctx.ISendrequestToWithBody)
	scenarioContext.Step(`^the response code should be (\d+)$`, ctx.TheResponseCodeShouldBe)
	scenarioContext.Step(`^the response body should be, in JSON:$`, ctx.TheResponseBodyShouldBeJSON)
	scenarioContext.Step(`^the response body should be:$`, ctx.TheResponseBodyShouldBe)
	scenarioContext.Step(`^the response body decoded as "([^"]+)" should be, in JSON:$`, ctx.TheResponseDecodedBodyShouldBeJSON)
	scenarioContext.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, ctx.TheResponseHeaderShouldBe)
	scenarioContext.Step(`^the response header "([^"]*)" should not be set$`, ctx.TheResponseHeaderShouldNotBeSet)
	scenarioContext.Step(`^the response headers? "([^"]*)" should be:`, ctx.TheResponseHeadersShouldBe)
	scenarioContext.Step(`^the response should be "([^"]*)"$`, ctx.TheResponseShouldBe)
	scenarioContext.Step(`^the response error message should contain "(.*)"$`, ctx.TheResponseErrorMessageShouldContain)

	scenarioContext.Step(`^the response should be a JSON array with (\d+) entr(?:ies|y)$`, ctx.ItShouldBeAJSONArrayWithEntries)
	scenarioContext.Step(`^the response at ([^ ]+) should be "([^"]*)"$`, ctx.TheResponseAtShouldBeTheValue)
	scenarioContext.Step("^the response at ([^ ]+) should be:$", ctx.TheResponseAtShouldBe)
	scenarioContext.Step("^the response at ([^ ]+) in JSON should be:$", ctx.TheResponseAtInJSONShouldBe)
	scenarioContext.Step("^the response at ([^ ]+) should be the base64 of an AES-256-GCM encrypted JSON object containing:",
		ctx.TheResponseAtShouldBeTheBase64OfAnAES256GCMEncryptedJSONObjectContaining)

	scenarioContext.Step(`^the table "([^"]*)" should be:$`, ctx.TableShouldBe)
	scenarioContext.Step(`^the table "([^"]*)" should be empty$`, ctx.TableShouldBeEmpty)
	scenarioContext.Step(`^the table "([^"]*)" at (\w+?)s? "([^"]*)" should be:$`, ctx.TableAtColumnValueShouldBe)
	scenarioContext.Step(`^the table "([^"]*)" at (\w+?)s? "([^"]*)" should be empty$`, ctx.TableAtColumnValueShouldBeEmpty)
	scenarioContext.Step(`^the table "([^"]*)" should not contain (\w+?)s? "([^"]*)"$`, ctx.TableShouldNotContainColumnValue)
	scenarioContext.Step(`^the table "([^"]*)" should remain unchanged$`, ctx.TableShouldRemainUnchanged)
	scenarioContext.Step(`^the table "([^"]*)" should remain unchanged, regardless of the rows? with (\w+?)s? "([^"]*)"$`,
		ctx.TableShouldStayUnchangedRegardlessOfTheRowsWithColumnValue)
	scenarioContext.Step(`^the table "([^"]*)" should remain unchanged, except that the rows? with ([^"]*) "([^"]*)" should be deleted$`,
		ctx.TableShouldRemainUnchangedExceptThatTheRowsWithColumnValueShouldBeDeleted)
	scenarioContext.Step(`^the DB time now is "([^"]*)"$`, ctx.DBTimeNow)

	scenarioContext.Step(`^"([^"]+)" is a token signed by (.+) with the following payload:$`, ctx.SignedTokenIsDistributed)
	scenarioContext.Step(`^"([^"]+)" is a falsified token signed by (.+) with the following payload:$`, ctx.FalsifiedSignedTokenIsDistributed)
	scenarioContext.Step(`^logs should contain:$`, ctx.LogsShouldContain)

	scenarioContext.Step(`^the login module "token" endpoint for code "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForCodeReturns)
	scenarioContext.Step(`^the login module "token" endpoint for code "([^"]*)" and code_verifier "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForCodeAndCodeVerifierReturns)
	scenarioContext.Step(
		`^the login module "token" endpoint for code "([^"]*)" and code_verifier "([^"]*)" and redirect_uri "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForCodeAndCodeVerifierAndRedirectURIReturns)
	scenarioContext.Step(`^the login module "token" endpoint for refresh token "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleTokenEndpointForRefreshTokenReturns)
	scenarioContext.Step(`^the login module "account" endpoint for token "([^"]*)" returns (\d+) with body:$`,
		ctx.TheLoginModuleAccountEndpointForTokenReturns)
	scenarioContext.Step(`^the login module "unlink_client" endpoint for user id "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleUnlinkClientEndpointForUserIDReturns)
	scenarioContext.Step(`^the login module "create" endpoint with params "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleCreateEndpointWithParamsReturns)
	scenarioContext.Step(`^the login module "delete" endpoint with params "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleDeleteEndpointWithParamsReturns)
	scenarioContext.Step(
		`^the login module "lti_result/send" endpoint for user id "([^"]*)", `+
			`content id "([^"]*)", score "([^"]*)" returns (\d+) with encoded body:$`,
		ctx.TheLoginModuleLTIResultSendEndpointForUserIDContentIDScoreReturns)

	scenarioContext.After(func(contextCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		tearDownErr := ctx.ScenarioTeardown(sc, err)
		if err == nil {
			err = tearDownErr
		}
		if restoreFunc != nil { // If we captured the output, restore it
			restoreFunc(err != nil) // Pass through the output if the test failed
		}
		if err != nil { // If the test failed, restore the output of the parent test (if it is captured)
			if parentOutputRestorerFunc := contextCtx.Value(outputRestorerFuncKey).(*flowmingo.RestoreFunc); *parentOutputRestorerFunc != nil {
				(*parentOutputRestorerFunc)(true)
				*parentOutputRestorerFunc = nil
			}

			fmt.Printf("\nLogs: \n%s\n", ctx.logsHook.GetAllStructuredLogs()) //nolint:forbidigo // Print the app logs of the failed test
		}
		return contextCtx, tearDownErr //nolint:nilnil // It looks like we really want to return the context even if there is an error.
	})
}
