//go:build !prod

package testhelpers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/utils"

	"github.com/jinzhu/gorm"

	"github.com/cucumber/godog"
)

// registerFeaturesForSessions registers the Gherkin features related to sessions and access tokens.
func (ctx *TestContext) registerFeaturesForSessions(s *godog.ScenarioContext) {
	s.Step(`^I am (@\w+)$`, ctx.IAm)
	s.Step(`^I am the user with id "([^"]*)"$`, ctx.IAmUserWithID)

	s.Step(`^there are the following sessions:$`, ctx.ThereAreTheFollowingSessions)
	s.Step(`^there are the following access tokens:$`, ctx.ThereAreTheFollowingAccessTokens)
	s.Step(`^there are (\d+) sessions for user (@\w+)$`, ctx.ThereAreCountSessionsForUser)
	s.Step(`^there is no session (@\w+)$`, ctx.ThereIsNoSessionID)
	s.Step(`^there are (\d+) access tokens for user (@\w+)$`, ctx.ThereAreCountAccessTokensForUser)
	s.Step(`^there is no access token "([^"]*)"$`, ctx.ThereIsNoAccessToken)
}

// addSession adds a session in database.
func (ctx *TestContext) addSession(session, user, refreshToken string) {
	sessionID := ctx.getReference(session)
	userID := ctx.getReference(user)

	ctx.addInDatabase("sessions", strconv.FormatInt(sessionID, 10), map[string]interface{}{
		"session_id":    sessionID,
		"user_id":       userID,
		"refresh_token": refreshToken,
	})
}

// addAccessToken adds an access token in database.
func (ctx *TestContext) addAccessToken(session, token, issuedAt, expiresAt string) {
	sessionID := ctx.getReference(session)

	issuedAtDate, err := time.Parse(utils.DateTimeFormat, issuedAt)
	if err != nil {
		panic(err)
	}

	expiresAtDate, err := time.Parse(utils.DateTimeFormat, expiresAt)
	if err != nil {
		panic(err)
	}

	ctx.addInDatabase("access_tokens", token, map[string]interface{}{
		"session_id": sessionID,
		"token":      token,
		"issued_at":  issuedAtDate,
		"expires_at": expiresAtDate,
	})
}

// IAm Sets the current user.
func (ctx *TestContext) IAm(name string) error {
	err := ctx.ThereIsAUser(name)
	if err != nil {
		return err
	}

	err = ctx.IAmUserWithID(ctx.getReference(name))
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

	db, err := database.Open(ctx.db)
	if err != nil {
		return err
	}
	return database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		err = store.Sessions().InsertMap(map[string]interface{}{
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
	userID := ctx.getReference(user)

	var sessionCount int
	err := ctx.db.QueryRow("SELECT COUNT(*) as count FROM sessions WHERE user_id = ?", userID).
		Scan(&sessionCount)
	if err != nil {
		return err
	}

	if sessionCount != count {
		return fmt.Errorf("expected %d sessions for user %s, got %d", count, user, sessionCount)
	}

	return nil
}

func (ctx *TestContext) ThereIsNoSessionID(session string) error {
	sessionID := ctx.getReference(session)

	var sessionCount int
	err := ctx.db.QueryRow("SELECT COUNT(*) as count FROM sessions WHERE session_id = ?", sessionID).
		Scan(&sessionCount)
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
	userID := ctx.getReference(user)

	var accessTokensCount int
	err := ctx.db.QueryRow(`
		SELECT COUNT(*) as count FROM access_tokens
			JOIN sessions ON sessions.session_id = access_tokens.session_id
		 WHERE sessions.user_id = ?`, userID).
		Scan(&accessTokensCount)
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
	err := ctx.db.QueryRow("SELECT COUNT(*) as count FROM access_tokens WHERE token = ?", accessToken).
		Scan(&accessTokensCount)
	if err != nil {
		return err
	}

	if accessTokensCount > 0 {
		return fmt.Errorf("there should be no access token %s", accessToken)
	}

	return nil
}
