//go:build !prod && !unit

package testhelpers

import (
	"strconv"

	"github.com/cucumber/godog"
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
	userGroupID := ctx.getIDOrIDByReference(user)
	primaryKey := ctx.getUserPrimaryKey(userGroupID)

	if !ctx.isInDatabase("users", primaryKey) {
		err := ctx.DBHasTable("users",
			constructGodogTableFromData([]stringKeyValuePair{
				{"group_id", strconv.FormatInt(userGroupID, 10)},
				{"login", referenceToName(user)},
				{"login_id", "null"},
				{"temp_user", "false"},
				{"first_name", "null"},
				{"last_name", "null"},
			}))
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

		groupID := ctx.getIDOrIDByReference(user["user"])

		err := ctx.ThereIsAUser(user["user"])
		mustNotBeError(err)

		primaryKey := ctx.getUserPrimaryKey(groupID)

		for key, value := range user {
			if key == "user" {
				continue
			}
			ctx.setUserFieldInDatabase(primaryKey, key, value)
		}
	}

	return nil
}
