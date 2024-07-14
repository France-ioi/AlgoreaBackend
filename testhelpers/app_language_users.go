//go:build !prod

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

// addUsersIntoAllUsersGroup adds all users in the AllUsers group if it is defined.
func (ctx *TestContext) addUsersIntoAllUsersGroup() error {
	if ctx.allUsersGroup == "" {
		return nil
	}

	for userID := range ctx.dbTables["users"] {
		err := ctx.UserIsAMemberOfTheGroup(userID, ctx.allUsersGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

// getUserPrimaryKey returns the primary key of a group.
func (ctx *TestContext) getUserPrimaryKey(groupID int64) string {
	return strconv.FormatInt(groupID, 10)
}

// addUser adds a user in database.
func (ctx *TestContext) addUser(user string) {
	primaryKey := ctx.getUserPrimaryKey(ctx.getReference(user))

	if !ctx.isInDatabase("users", primaryKey) {
		ctx.addInDatabase("users", primaryKey, map[string]interface{}{
			"group_id": ctx.getReference(user),
			"login":    referenceToName(user),
			// All the other fields are set to default values.
			"login_id":   nil,
			"temp_user":  false,
			"first_name": nil,
			"last_name":  nil,
		})
	}
}

// setUserFieldInDatabase sets a specific field of a user in the database.
func (ctx *TestContext) setUserFieldInDatabase(primaryKey, field string, value interface{}) {
	if value == tableValueNull {
		value = nil
	}
	if value == tableValueFalse {
		value = false
	}
	if value == tableValueTrue {
		value = true
	}

	ctx.dbTables["users"][primaryKey][field] = value
}

// ThereIsAUser create a user.
func (ctx *TestContext) ThereIsAUser(name string) error {
	ctx.addUser(name)

	err := ctx.ThereIsAGroup(name)
	mustNotBeError(err)

	groupPrimaryKey := ctx.getGroupPrimaryKey(ctx.getReference(name))
	ctx.setGroupFieldInDatabase(groupPrimaryKey, "type", "User")

	return nil
}

// ThereAreTheFollowingUsers defines users.
func (ctx *TestContext) ThereAreTheFollowingUsers(users *godog.Table) error {
	for i := 1; i < len(users.Rows); i++ {
		user := ctx.getRowMap(i, users)

		groupID := ctx.getReference(user["user"])

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
