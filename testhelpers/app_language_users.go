//go:build !prod

package testhelpers

import (
	"strconv"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
)

// registerFeaturesForUsers registers the Gherkin features related to users.
func (ctx *TestContext) registerFeaturesForUsers(s *godog.ScenarioContext) {
	s.Step(`^there is a user (@\w+)$`, ctx.ThereIsAUser)
	s.Step(`^there are the following users:$`, ctx.ThereAreTheFollowingUsers)
}

// getUserPrimaryKey returns the primary key of a group.
func (ctx *TestContext) getUserPrimaryKey(groupID int64) map[string]string {
	return map[string]string{"group_id": strconv.FormatInt(groupID, 10)}
}

// addUser adds a user to the database.
func (ctx *TestContext) addUser(user string) {
	userGroupID := ctx.getIDOfReference(user)
	primaryKey := ctx.getUserPrimaryKey(userGroupID)

	if !ctx.isInDatabase("users", primaryKey) {
		err := ctx.DBHasTable("users", &godog.Table{
			Rows: []*messages.PickleTableRow{
				{Cells: []*messages.PickleTableCell{
					{Value: "group_id"},
					{Value: "login"},
					{Value: "login_id"},
					{Value: "temp_user"},
					{Value: "first_name"},
					{Value: "last_name"},
				}},
				{Cells: []*messages.PickleTableCell{
					{Value: strconv.FormatInt(userGroupID, 10)},
					{Value: referenceToName(user)},
					// All the other fields are set to default values.
					{Value: "null"},
					{Value: "false"},
					{Value: "null"},
					{Value: "null"},
				}},
			},
		})
		if err != nil {
			panic(err)
		}
	}
}

// setUserFieldInDatabase sets a specific field of a user in the database.
func (ctx *TestContext) setUserFieldInDatabase(primaryKey map[string]string, field, value string) {
	ctx.setDBTableRowColumnValue("users", primaryKey, field, value)
}

// ThereIsAUser create a user.
func (ctx *TestContext) ThereIsAUser(name string) (err error) {
	defer recoverPanics(&err)

	ctx.addUser(name)
	ctx.addGroup(name, "User")

	return nil
}

// ThereAreTheFollowingUsers defines users.
func (ctx *TestContext) ThereAreTheFollowingUsers(users *godog.Table) error {
	for i := 1; i < len(users.Rows); i++ {
		user := ctx.getRowMap(i, users)

		groupID := ctx.getIDOfReference(user["user"])

		err := ctx.ThereIsAUser(user["user"])
		mustNotBeError(err)

		if _, ok := user["first_name"]; ok {
			ctx.setUserFieldInDatabase(
				ctx.getUserPrimaryKey(groupID),
				"first_name",
				user["first_name"],
			)
		}
		if _, ok := user["last_name"]; ok {
			ctx.setUserFieldInDatabase(
				ctx.getUserPrimaryKey(groupID),
				"last_name",
				user["last_name"],
			)
		}
		if _, ok := user["temp_user"]; ok {
			ctx.setUserFieldInDatabase(
				ctx.getUserPrimaryKey(groupID),
				"temp_user",
				user["temp_user"],
			)
		}
		if _, ok := user["login_id"]; ok {
			ctx.setUserFieldInDatabase(
				ctx.getUserPrimaryKey(groupID),
				"login_id",
				user["login_id"],
			)
		}
	}

	return nil
}
