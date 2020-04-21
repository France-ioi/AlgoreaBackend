package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /auth/token auth accessTokenCreate
// ---
// summary: Create or refresh an access token
// description:
//     If the "Authorization" header is not given, the service converts the given OAuth2 authorization code into tokens,
//     creates or updates the authenticated user in the DB with the data returned by the login module,
//     and saves new access & refresh tokens into the DB as well.
//     If OAuth2 authentication has used the PKCE extension, the '{code_verifier}' should be provided
//     so it can be sent together with the '{code}' to the authentication server.
//
//
//     If the "Authorization" header is given, the service refreshes the access token
//     (locally for temporary users or via the login module for normal users) and
//     saves it into the DB keeping only the input token (from authorization headers) and the new token.
//     Since the login module responds with both access and refresh tokens, the service updates the user's
//     refresh token in this case as well.
//
//
//   * One of the “Authorization” header and the '{code}' parameter should be present (not both at once).
// security: []
// parameters:
// - name: code
//   in: query
//   description: OAuth2 code
//   type: string
// - name: code_verifier
//   in: query
//   description: OAuth2 PKCE code verifier
//   type: string
// responses:
//   "201":
//     description: "Created. Success response with the new access token"
//     in: body
//     schema:
//       "$ref": "#/definitions/userCreateTmpResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createAccessToken(w http.ResponseWriter, r *http.Request) service.APIError {
	// "Authorization" header is given, requesting a new token from the given token
	if len(r.Header["Authorization"]) != 0 {
		if len(r.URL.Query()["code"]) != 0 {
			return service.ErrInvalidRequest(
				errors.New("only one of the 'code' query parameter and the 'Authorization' header can be given"))
		}
		auth.UserMiddleware(srv.Store.Sessions())(service.AppHandler(srv.refreshAccessToken)).ServeHTTP(w, r)
		return service.NoError
	}

	// the code is given, requesting a token from code and optionally code_verifier, and create/update user.
	code, err := service.ResolveURLQueryGetStringField(r, "code")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	oauthConfig := auth.GetOAuthConfig(srv.AuthConfig)
	oauthOptions := make([]oauth2.AuthCodeOption, 0, 1)
	if len(r.URL.Query()["code_verifier"]) != 0 {
		oauthOptions = append(oauthOptions, oauth2.SetAuthURLParam("code_verifier", r.URL.Query().Get("code_verifier")))
	}

	token, err := oauthConfig.Exchange(r.Context(), code, oauthOptions...)
	service.MustNotBeError(err)

	userProfile, err := loginmodule.NewClient(srv.AuthConfig.GetString("LoginModuleURL")).GetUserProfile(r.Context(), token.AccessToken)
	service.MustNotBeError(err)
	userProfile["last_ip"] = strings.SplitN(r.RemoteAddr, ":", 2)[0]

	domainConfig := domain.ConfigFromContext(r.Context())

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		userID := createOrUpdateUser(store.Users(), userProfile, domainConfig)
		service.MustNotBeError(store.Sessions().InsertNewOAuth(userID, token))

		service.MustNotBeError(store.Exec(
			"INSERT INTO refresh_tokens (user_id, refresh_token) VALUES (?, ?) ON DUPLICATE KEY UPDATE refresh_token = ?",
			userID, token.RefreshToken, token.RefreshToken).Error())

		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": token.AccessToken,
		"expires_in":   time.Until(token.Expiry).Round(time.Second) / time.Second,
	})))
	return service.NoError
}

func createOrUpdateUser(s *database.UserStore, userData map[string]interface{}, domainConfig *domain.Configuration) int64 {
	var groupID int64
	err := s.WithWriteLock().
		Where("login_id = ?", userData["login_id"]).PluckFirst("group_id", &groupID).Error()

	userData["latest_login_at"] = database.Now()
	userData["latest_activity_at"] = database.Now()

	if defaultLanguage, ok := userData["default_language"]; ok && defaultLanguage == nil {
		userData["default_language"] = database.Default()
	}

	if gorm.IsRecordNotFoundError(err) {
		selfGroupID := createGroupsFromLogin(s.Groups(), userData["login"].(string), domainConfig)
		userData["temp_user"] = 0
		userData["registered_at"] = database.Now()
		userData["group_id"] = selfGroupID

		service.MustNotBeError(s.Users().InsertMap(userData))
		service.MustNotBeError(s.Attempts().InsertMap(map[string]interface{}{
			"participant_id": selfGroupID,
			"id":             0,
			"creator_id":     selfGroupID,
			"created_at":     database.Now(),
		}))

		return selfGroupID
	}

	found, err := s.GroupGroups().WithWriteLock().Where("parent_group_id = ?", domainConfig.RootSelfGroupID).
		Where("child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	groupsToCreate := make([]map[string]interface{}, 0, 2)
	if !found {
		groupsToCreate = append(groupsToCreate,
			map[string]interface{}{"parent_group_id": domainConfig.RootSelfGroupID, "child_group_id": groupID})
	}

	service.MustNotBeError(s.GroupGroups().CreateRelationsWithoutChecking(groupsToCreate))

	service.MustNotBeError(err)
	service.MustNotBeError(s.ByID(groupID).UpdateColumn(userData).Error())
	return groupID
}

func createGroupsFromLogin(store *database.GroupStore, login string, domainConfig *domain.Configuration) (selfGroupID int64) {
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
		selfGroupID = retryIDStore.NewID()
		return retryIDStore.Groups().InsertMap(map[string]interface{}{
			"id":          selfGroupID,
			"name":        login,
			"type":        "User",
			"description": login,
			"created_at":  database.Now(),
			"is_open":     false,
			"send_emails": false,
		})
	}))

	service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking([]map[string]interface{}{
		{"parent_group_id": domainConfig.RootSelfGroupID, "child_group_id": selfGroupID},
	}))

	return selfGroupID
}
