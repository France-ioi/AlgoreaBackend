//go:build !prod && !unit

package testhelpers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

// registerFeaturesForSessions registers the Gherkin features related to sessions and access tokens.
func (ctx *TestContext) registerFeaturesForSessions(scenarioContext *godog.ScenarioContext) {
	scenarioContext.Step(`^I am (@\w+)$`, ctx.IAm)
	scenarioContext.Step(`^I am the user with id "([^"]*)"$`, ctx.IAmUserWithID)

	scenarioContext.Step(`^there are the following sessions:$`, ctx.ThereAreTheFollowingSessions)
	scenarioContext.Step(`^there are the following access tokens:$`, ctx.ThereAreTheFollowingAccessTokens)
	scenarioContext.Step(`^there are (\d+) sessions for user (@\w+)$`, ctx.ThereAreCountSessionsForUser)
	scenarioContext.Step(`^there is no session (@\w+)$`, ctx.ThereIsNoSessionID)
	scenarioContext.Step(`^there are (\d+) access tokens for user (@\w+)$`, ctx.ThereAreCountAccessTokensForUser)
	scenarioContext.Step(`^there is no access token "([^"]*)"$`, ctx.ThereIsNoAccessToken)
}

// addSession adds a session to the database.
func (ctx *TestContext) addSession(session, user, refreshToken string) {
	sessionID := ctx.getIDOrIDByReference(session)
	userID := ctx.getIDOrIDByReference(user)

	err := ctx.DBHasTable("sessions",
		constructGodogTableFromData([]stringKeyValuePair{
			{"session_id", strconv.FormatInt(sessionID, 10)},
			{"user_id", strconv.FormatInt(userID, 10)},
			{"refresh_token", refreshToken},
		}))
	if err != nil {
		panic(err)
	}
}

// addAccessToken adds an access token to the database.
func (ctx *TestContext) addAccessToken(session, token, issuedAt, expiresAt string) {
	sessionID := ctx.getIDOrIDByReference(session)

	_, err := time.Parse(time.DateTime, issuedAt)
	if err != nil {
		panic(err)
	}

	_, err = time.Parse(time.DateTime, expiresAt)
	if err != nil {
		panic(err)
	}

	err = ctx.DBHasTable("access_tokens",
		constructGodogTableFromData([]stringKeyValuePair{
			{"session_id", strconv.FormatInt(sessionID, 10)},
			{"token", token},
			{"issued_at", issuedAt},
			{"expires_at", expiresAt},
		}))
	if err != nil {
		panic(err)
	}
}

// IAm Sets the current user.
func (ctx *TestContext) IAm(name string) error {
	err := ctx.ThereIsAUser(name)
	if err != nil {
		return err
	}

	err = ctx.IAmUserWithID(ctx.getIDOrIDByReference(name))
	if err != nil {
		return err
	}

	ctx.user = name

	return nil
}

// IAmUserWithID sets the current logged user to the one with the provided ID.
func (ctx *TestContext) IAmUserWithID(userID int64) error {
	ctx.userID = userID
	ctx.user = strconv.FormatInt(userID, 10)

	return database.NewDataStore(ctx.application.Database).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		err := store.Sessions().InsertMap(map[string]interface{}{
			"session_id": testSessionID,
			"user_id":    ctx.userID,
		})
		if err != nil {
			return err
		}

		return store.AccessTokens().InsertMap(map[string]interface{}{
			"session_id": testSessionID,
			"token":      testAccessToken,
			"issued_at":  database.Now(),
			"expires_at": gorm.Expr("? + INTERVAL 7200 SECOND", database.Now()),
		})
	})
}

// ThereAreTheFollowingSessions create sessions.
func (ctx *TestContext) ThereAreTheFollowingSessions(sessions *godog.Table) error {
	for i := 1; i < len(sessions.Rows); i++ {
		session := ctx.getRowMap(i, sessions)

		ctx.addSession(
			session["session"],
			session["user"],
			session["refresh_token"],
		)
	}

	return nil
}

// ThereAreCountSessionsForUser checks if there are a given number of sessions for a given user.
func (ctx *TestContext) ThereAreCountSessionsForUser(count int, user string) error {
	userID := ctx.getIDOrIDByReference(user)

	var sessionCount int
	err := ctx.application.Database.Table("sessions").Where("user_id = ?", userID).Count(&sessionCount).Error()
	if err != nil {
		return err
	}

	if sessionCount != count {
		return fmt.Errorf("expected %d sessions for user %s, got %d", count, user, sessionCount)
	}

	return nil
}

// ThereIsNoSessionID checks that a session with given session ID doesn't exist.
func (ctx *TestContext) ThereIsNoSessionID(session string) error {
	sessionID := ctx.getIDOrIDByReference(session)

	var sessionCount int
	err := ctx.application.Database.Table("sessions").Where("session_id = ?", sessionID).Count(&sessionCount).Error()
	if err != nil {
		return err
	}

	if sessionCount > 0 {
		return fmt.Errorf("there should be no session with ID %d", sessionID)
	}

	return nil
}

// ThereAreTheFollowingAccessTokens create access tokens.
func (ctx *TestContext) ThereAreTheFollowingAccessTokens(accessTokens *godog.Table) error {
	for i := 1; i < len(accessTokens.Rows); i++ {
		accessToken := ctx.getRowMap(i, accessTokens)

		ctx.addAccessToken(
			accessToken["session"],
			accessToken["token"],
			accessToken["issued_at"],
			accessToken["expires_at"],
		)
	}

	return nil
}

// ThereAreCountAccessTokensForUser checks if there are a given number of access tokens for a given user.
func (ctx *TestContext) ThereAreCountAccessTokensForUser(count int, user string) error {
	userID := ctx.getIDOrIDByReference(user)

	var accessTokensCount int
	err := ctx.application.Database.Table("access_tokens").
		Joins("JOIN sessions ON sessions.session_id = access_tokens.session_id").
		Where(`sessions.user_id = ?`, userID).
		Count(&accessTokensCount).Error()
	if err != nil {
		return err
	}

	if accessTokensCount != count {
		return fmt.Errorf("expected %d access tokens for user %s, got %d", count, user, accessTokensCount)
	}

	return nil
}

// ThereIsNoAccessToken checks that an access token doesn't exist.
func (ctx *TestContext) ThereIsNoAccessToken(accessToken string) error {
	var accessTokensCount int
	err := ctx.application.Database.Table("access_tokens").Where("token = ?", accessToken).Count(&accessTokensCount).Error()
	if err != nil {
		return err
	}

	if accessTokensCount > 0 {
		return fmt.Errorf("there should be no access token %s", accessToken)
	}

	return nil
}
