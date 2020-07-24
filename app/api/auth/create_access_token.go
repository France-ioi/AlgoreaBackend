package auth

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
//     If OAuth2 authentication has used the PKCE extension, the `{code_verifier}` should be provided
//     so it can be sent together with the `{code}` to the authentication server.
//
//
//     If the "Authorization" header is given, the service refreshes the access token
//     (locally for temporary users or via the login module for normal users) and
//     saves it into the DB keeping only the input token (from authorization headers) and the new token.
//     Since the login module responds with both access and refresh tokens, the service updates the user's
//     refresh token in this case as well. If there is no refresh token for the user in the DB,
//     the 'not found' error is returned.
//
//
//   * One of the “Authorization” header and the `{code}` parameter should be present (not both at once).
// security: []
// consumes:
//   - application/json
//   - application/x-www-form-urlencoded
// parameters:
// - name: code
//   in: query
//   description: OAuth2 code (can also be given in form data)
//   type: string
// - name: code_verifier
//   in: query
//   description: OAuth2 PKCE code verifier  (can also be given in form data)
//   type: string
// - name: redirect_uri
//   in: query
//   description: OAuth2 redirection URI
//   type: string
// - in: body
//   name: parameters
//   description: The optional parameters can be given in the body as well
//   schema:
//     type: object
//     properties:
//       code:
//         type: string
//         description: OAuth2 code
//       code_verifier:
//         type: string
//         description: OAuth2 PKCE code verifier
//       redirect_uri:
//         type: string
//         description: OAuth2 redirection URI
// responses:
//   "201":
//     description: "Created. Success response with the new access token"
//     in: body
//     schema:
//       "$ref": "#/definitions/userCreateTmpResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createAccessToken(w http.ResponseWriter, r *http.Request) service.APIError {
	requestData, apiError := parseRequestParametersForCreateAccessToken(r)
	if apiError != service.NoError {
		return apiError
	}

	// "Authorization" header is given, requesting a new token from the given token
	if len(r.Header["Authorization"]) != 0 {
		if _, ok := requestData["code"]; ok {
			return service.ErrInvalidRequest(
				errors.New("only one of the 'code' parameter and the 'Authorization' header can be given"))
		}
		auth.UserMiddleware(srv.Store.Sessions())(service.AppHandler(srv.refreshAccessToken)).ServeHTTP(w, r)
		return service.NoError
	}

	// the code is given, requesting a token from code and optionally code_verifier, and create/update user.
	code, ok := requestData["code"]
	if !ok {
		return service.ErrInvalidRequest(errors.New("missing code"))
	}

	oauthConfig := auth.GetOAuthConfig(srv.AuthConfig)
	oauthOptions := make([]oauth2.AuthCodeOption, 0, 1)
	if codeVerifier, ok := requestData["code_verifier"]; ok {
		oauthOptions = append(oauthOptions, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
	if redirectURI, ok := requestData["redirect_uri"]; ok {
		oauthOptions = append(oauthOptions, oauth2.SetAuthURLParam("redirect_uri", redirectURI))
	}

	token, err := oauthConfig.Exchange(r.Context(), code, oauthOptions...)
	service.MustNotBeError(err)

	userProfile, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).GetUserProfile(r.Context(), token.AccessToken)
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

func parseRequestParametersForCreateAccessToken(r *http.Request) (map[string]string, service.APIError) {
	requestData := make(map[string]string, 2)
	query := r.URL.Query()
	extractOptionalParameter(query, "code", requestData)
	extractOptionalParameter(query, "code_verifier", requestData)
	extractOptionalParameter(query, "redirect_uri", requestData)

	contentType := strings.ToLower(strings.TrimSpace(strings.SplitN(r.Header.Get("Content-Type"), ";", 2)[0]))
	switch contentType {
	case "application/json":
		var jsonPayload struct {
			Code         *string `json:"code"`
			CodeVerifier *string `json:"code_verifier"`
			RedirectURI  *string `json:"redirect_uri"`
		}
		defer func() { _, _ = io.Copy(ioutil.Discard, r.Body) }()
		err := json.NewDecoder(r.Body).Decode(&jsonPayload)
		if err != nil {
			return nil, service.ErrInvalidRequest(err)
		}
		if jsonPayload.Code != nil {
			requestData["code"] = *jsonPayload.Code
		}
		if jsonPayload.CodeVerifier != nil {
			requestData["code_verifier"] = *jsonPayload.CodeVerifier
		}
		if jsonPayload.RedirectURI != nil {
			requestData["redirect_uri"] = *jsonPayload.RedirectURI
		}
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			return nil, service.ErrInvalidRequest(err)
		}
		extractOptionalParameter(r.PostForm, "code", requestData)
		extractOptionalParameter(r.PostForm, "code_verifier", requestData)
		extractOptionalParameter(r.PostForm, "redirect_uri", requestData)
	}
	return requestData, service.NoError
}

func extractOptionalParameter(query url.Values, paramName string, requestData map[string]string) {
	if len(query[paramName]) != 0 {
		requestData[paramName] = query.Get(paramName)
	}
}

func createOrUpdateUser(s *database.UserStore, userData map[string]interface{}, domainConfig *domain.CtxConfig) int64 {
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
	service.MustNotBeError(err)

	found, err := s.GroupGroups().WithWriteLock().Where("parent_group_id = ?", domainConfig.RootSelfGroupID).
		Where("child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	groupsToCreate := make([]map[string]interface{}, 0, 2)
	if !found {
		groupsToCreate = append(groupsToCreate,
			map[string]interface{}{"parent_group_id": domainConfig.RootSelfGroupID, "child_group_id": groupID})
	}

	service.MustNotBeError(s.GroupGroups().CreateRelationsWithoutChecking(groupsToCreate))
	service.MustNotBeError(s.ByID(groupID).UpdateColumn(userData).Error())
	return groupID
}

func createGroupsFromLogin(store *database.GroupStore, login string, domainConfig *domain.CtxConfig) (selfGroupID int64) {
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
