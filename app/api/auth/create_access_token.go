package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

type ctxKey int

const (
	parsedRequestData             ctxKey = iota
	maxNumberOfUserSessionsToKeep        = 10
)

// swagger:operation POST /auth/token auth accessTokenCreate
//
//	---
//	summary: Create or refresh an access token, may create a temporary user
//	description:
//
//		This service is called by the frontend in the following contexts
//			- After the user successfully logs-in on the login-module
//			- When the frontend loads, to verify if the user is already logged-in,
//				because if token are only stored in cookies, the frontend never has the hand on it and so does not know
//				on launch whether the user is logged or not.
//				If the user is not already logged, a temporary user is created.
//
//
//		To avoid the spamming of the sessions table with session creation, we store a maximum of 10 sessions per user.
//		When we reach this limit, we delete the oldest session of the user.
//
//
//		The `{code}` parameter is an output of the login-module after a successful login.
//
//		* If the `{code}` is given and the "Authorization" header is not given.
//			This happens after the client successfully logged on the login-module.
//			Then,
//			the service converts the given OAuth2 authorization code into tokens,
//			creates or updates the authenticated user in the DB with the data returned by the login module,
//			and saves new access & refresh tokens into the DB as well.
//			If OAuth2 authentication has used the PKCE extension, the `{code_verifier}` should be provided
//			so it can be sent together with the `{code}` to the authentication server.
//
//
//		* If the `{code}` is not given while the "Authorization" header or/and the "access_token" cookie is given
//			(when both are given, the "Authorization" header is used, and the cookie gets deleted),
//			and if the authentication is valid.
//			This happens when the frontend app loads and the user is already logged-in.
//			Then,
//
//
//			1. If the access token used isn't the most recent access token of the user, we return the most recent access token.
//
//
//			2. If the access token used is the most recent access token of the user, and it has been refreshed AFTER 5 minutes ago,
//				we just return it.
//
//
//			3. If the access token used is the most recent access token of the user, and it has been refreshed BEFORE 5 minutes ago,
//				we refresh the access token and return the new access token
//				(locally for temporary users or via the login module for normal users) and
//				saves it into the DB keeping only the input token and the new token.
//				Since the login module responds with both access and refresh tokens, the service updates the user's
//				refresh token in this case as well.
//				We also delete all the expired tokens of the user to keep the database leaner.
//				If there is no refresh token for the user in the DB or if the refresh token has expired,
//				the 'not found' error is returned.
//
//
//		* If the `{code}` is not given,
//			and if the authentication is not given (no "Authorization" header and no "access_token" cookie) or is invalid.
//			This happens when the frontend app loads, and the user is not logged-in, or if the authentication
//			is not valid anymore.
//			Then,
//			if `create_temp_user_if_not_authorized`=`true`,
//			we create a temporary user and log-in the user as it.
//			This happens from the frontend when a user that was once logged-in comes back to the website after the token expired.
//			Otherwise,
//			if `create_temp_user_if_not_authorized`=`false`,
//			an error is returned.
//			This happens from the frontend when a user that is logged-in becomes inactive for a while, while the tab is	still open.
//			When the tab is then restored, for example, after the computer comes back from sleep,
//			it tries to reconnect but the token has expired.
//			Note, when the tab is open and active, the frontend automatically refreshes the token before it expires.
//
//
//		If attributes of the old and the new 'access_token' cookies are different (or the token is returned in the JSON),
//		the old cookie gets deleted (otherwise, just overwritten).
//
//		If a cookie is given together with a `{code}`, the cookie is deleted.
//
//		`{default_language}` is used only if a temporary user is created.
//		If it is not provided, the `DEFAULT` definition of `default_language` in the `users` table is used.
//
//
//		Validations
//			* The "Authorization" header is not allowed when the `{code}` is given.
//
//			* When `{use_cookie}`=1, at least one of `{cookie_secure}` and `{cookie_same_site}` must be true.
//	security: []
//	consumes:
//		- application/json
//		- application/x-www-form-urlencoded
//	parameters:
//		- name: code
//			in: query
//			description: OAuth2 code (can also be given in form data)
//			type: string
//		- name: code_verifier
//			in: query
//			description: OAuth2 PKCE code verifier  (can also be given in form data)
//			type: string
//		- name: redirect_uri
//			in: query
//			description: OAuth2 redirection URI
//			type: string
//		- name: use_cookie
//			in: query
//			description: If 1, set a cookie instead of returning the OAuth2 code in the data
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: cookie_secure
//			in: query
//			description: If 1, set the cookie with the `Secure` attribute
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: cookie_same_site
//			in: query
//			description: If 1, set the cookie with the `SameSite`='Strict' attribute value and with `SameSite`='None' otherwise
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: create_temp_user_if_not_authorized
//			description: Whether to create a temporary user if the token is not provided or expired.
//			in: query
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: default_language
//			description: The temporary user's default language.	Only if `create_temp_user_if_not_authorized`=`true`.
//			in: query
//			type: string
//			maxLength: 3
//		- in: body
//			name: parameters
//			description: The optional parameters can be given in the body as well
//			schema:
//				type: object
//				properties:
//					code:
//						type: string
//						description: OAuth2 code
//					code_verifier:
//						type: string
//						description: OAuth2 PKCE code verifier
//					redirect_uri:
//						type: string
//						description: OAuth2 redirection URI
//					use_cookie:
//						type: boolean
//						description: If true, set a cookie instead of returning the OAuth2 code in the data
//					cookie_secure:
//						type: boolean
//						description: If true, set the cookie with the `Secure` attribute
//					cookie_same_site:
//						type: boolean
//						description: If true, set the cookie with the `SameSite`='Strict' attribute value and with `SameSite`='None' otherwise
//	responses:
//		"201":
//			description: >
//				Created.
//				Success response with the new access token"
//			in: body
//			schema:
//				"$ref": "#/definitions/userCreateTmpResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createAccessToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	requestData, err := parseRequestParametersForCreateAccessToken(httpRequest)
	service.MustNotBeError(err)

	cookieAttributes, err := srv.resolveCookieAttributes(httpRequest, requestData)
	service.MustNotBeError(err)

	code, codeGiven := requestData["code"]
	if codeGiven && len(httpRequest.Header["Authorization"]) != 0 {
		return service.ErrInvalidRequest(
			errors.New("only one of the 'code' parameter and the 'Authorization' header can be given"))
	}

	if !codeGiven {
		var (
			requestContext context.Context
			authorized     bool
			reason         string
		)
		// The code is not given, requesting a new token from the given token.
		requestContext, authorized, reason, err = auth.ValidatesUserAuthentication(srv.Base, responseWriter, httpRequest)
		service.MustNotBeError(err)

		if authorized {
			service.AppHandler(srv.refreshAccessToken).
				ServeHTTP(responseWriter, httpRequest.WithContext(context.WithValue(requestContext, parsedRequestData, requestData)))
			return nil
		}

		var createTempUser bool
		createTempUser, err = service.ResolveURLQueryGetBoolFieldWithDefault(httpRequest, "create_temp_user_if_not_authorized", false)
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		if !createTempUser {
			return &service.APIError{
				HTTPStatusCode: http.StatusUnauthorized,
				EmbeddedError:  errors.New(reason),
			}
		}

		// createTempUser checks that the Authorization header is not present.
		// But from here, we want to be able to create a temporary user if the authorization is invalid,
		// for example, because it expired.
		// Since we don't need its value anymore, we just delete it.
		httpRequest.Header.Del("Authorization")

		service.AppHandler(srv.createTempUser).ServeHTTP(responseWriter, httpRequest)

		return nil
	}

	oauthConfig := auth.GetOAuthConfig(srv.AuthConfig)
	oauthOptions := make([]oauth2.AuthCodeOption, 0, 1)
	if codeVerifier, ok := requestData["code_verifier"]; ok {
		oauthOptions = append(oauthOptions, oauth2.SetAuthURLParam("code_verifier", codeVerifier.(string)))
	}
	if redirectURI, ok := requestData["redirect_uri"]; ok {
		oauthOptions = append(oauthOptions, oauth2.SetAuthURLParam("redirect_uri", redirectURI.(string)))
	}

	token, err := oauthConfig.Exchange(httpRequest.Context(), code.(string), oauthOptions...)
	service.MustNotBeError(err)

	expiresIn, err := validateAndGetExpiresInFromOAuth2Token(token)
	service.MustNotBeError(err)

	userProfile, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).
		GetUserProfile(httpRequest.Context(), token.AccessToken)
	service.MustNotBeError(err)
	userProfile["last_ip"] = strings.SplitN(httpRequest.RemoteAddr, ":", 2)[0] //nolint:mnd // cut off the port

	domainConfig := domain.ConfigFromContext(httpRequest.Context())

	service.MustNotBeError(srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		userID := createOrUpdateUser(store.Users(), userProfile, domainConfig)
		logging.LogEntrySetField(httpRequest, "user_id", userID)
		service.MustNotBeError(store.Groups().StoreBadges(userProfile["badges"].([]database.Badge), userID, true))

		sessionID := rand.Int63()
		service.MustNotBeError(store.Exec(
			"INSERT INTO sessions (session_id, user_id, refresh_token) VALUES (?, ?, ?)",
			sessionID, userID, token.RefreshToken).Error())
		service.MustNotBeError(store.AccessTokens().InsertNewToken(sessionID, token.AccessToken, expiresIn))

		// Delete the oldest sessions of the user keeping up to maxNumberOfUserSessionsToKeep sessions.
		store.Sessions().DeleteOldSessionsToKeepMaximum(userID, maxNumberOfUserSessionsToKeep)

		return nil
	}))

	srv.respondWithNewAccessToken(
		responseWriter, httpRequest, service.CreationSuccess[map[string]interface{}], token.AccessToken, expiresIn, cookieAttributes)
	return nil
}

func (srv *Service) respondWithNewAccessToken(responseWriter http.ResponseWriter, httpRequest *http.Request,
	rendererGenerator func(map[string]interface{}) render.Renderer, token string, expiresIn int32,
	cookieAttributes *auth.SessionCookieAttributes,
) {
	response := map[string]interface{}{
		"expires_in": expiresIn,
	}
	oldCookieAttributes := auth.SessionCookieAttributesFromContext(httpRequest.Context())
	if oldCookieAttributes == nil {
		oldCookieAttributes = &auth.SessionCookieAttributes{}
		_, *oldCookieAttributes = auth.ParseSessionCookie(httpRequest)
	}
	if oldCookieAttributes != nil && oldCookieAttributes.UseCookie && *oldCookieAttributes != *cookieAttributes {
		http.SetCookie(responseWriter, oldCookieAttributes.SessionCookie("", -1000))
	}
	if cookieAttributes.UseCookie {
		http.SetCookie(responseWriter, cookieAttributes.SessionCookie(token, expiresIn))
	} else {
		response["access_token"] = token
	}
	service.MustNotBeError(render.Render(responseWriter, httpRequest, rendererGenerator(response)))
}

func (srv *Service) resolveCookieAttributes(httpRequest *http.Request, requestData map[string]interface{}) (
	cookieAttributes *auth.SessionCookieAttributes, err error,
) {
	cookieAttributes = &auth.SessionCookieAttributes{}
	if value, ok := requestData["use_cookie"]; ok && value.(bool) {
		cookieAttributes.UseCookie = true
		cookieAttributes.Domain = domain.CurrentDomainFromContext(httpRequest.Context())
		cookieAttributes.Path = srv.ServerConfig.GetString("rootPath")
		if value, ok := requestData["cookie_secure"]; ok && value.(bool) {
			cookieAttributes.Secure = true
		}
		if value, ok := requestData["cookie_same_site"]; ok && value.(bool) {
			cookieAttributes.SameSite = true
		}
		if !cookieAttributes.Secure && !cookieAttributes.SameSite {
			return nil, service.ErrInvalidRequest(errors.New("one of cookie_secure and cookie_same_site must be true when use_cookie is true"))
		}
	}
	return cookieAttributes, nil
}

func parseRequestParametersForCreateAccessToken(httpRequest *http.Request) (map[string]interface{}, error) {
	allowedParameters := []string{
		"code", "code_verifier", "redirect_uri",
		"use_cookie", "cookie_secure", "cookie_same_site",
	}
	requestData := make(map[string]interface{}, len(allowedParameters))
	query := httpRequest.URL.Query()
	for _, parameterName := range allowedParameters {
		extractOptionalParameter(query, parameterName, requestData)
	}

	contentType := strings.ToLower(strings.TrimSpace(
		strings.SplitN(httpRequest.Header.Get("Content-Type"), ";", 2)[0])) //nolint:mnd // cut off the parameters, keep only the media type
	switch contentType {
	case "application/json":
		if err := parseJSONParams(httpRequest, requestData); err != nil {
			return nil, err
		}
	case "application/x-www-form-urlencoded":
		err := httpRequest.ParseForm()
		if err != nil {
			return nil, service.ErrInvalidRequest(err)
		}
		for _, parameterName := range allowedParameters {
			extractOptionalParameter(httpRequest.PostForm, parameterName, requestData)
		}
	}
	return preprocessBooleanCookieAttributes(requestData)
}

func parseJSONParams(httpRequest *http.Request, requestData map[string]interface{}) error {
	var jsonPayload struct {
		Code           *string `json:"code"`
		CodeVerifier   *string `json:"code_verifier"`
		RedirectURI    *string `json:"redirect_uri"`
		UseCookie      *bool   `json:"use_cookie"`
		CookieSecure   *bool   `json:"cookie_secure"`
		CookieSameSite *bool   `json:"cookie_same_site"`
	}
	defer func() { _, _ = io.Copy(io.Discard, httpRequest.Body) }()
	err := json.NewDecoder(httpRequest.Body).Decode(&jsonPayload)
	if err != nil {
		return service.ErrInvalidRequest(err)
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
	bool2String := map[bool]string{false: "0", true: "1"}
	if jsonPayload.UseCookie != nil {
		requestData["use_cookie"] = bool2String[*jsonPayload.UseCookie]
	}
	if jsonPayload.CookieSecure != nil {
		requestData["cookie_secure"] = bool2String[*jsonPayload.CookieSecure]
	}
	if jsonPayload.CookieSameSite != nil {
		requestData["cookie_same_site"] = bool2String[*jsonPayload.CookieSameSite]
	}
	return nil
}

func preprocessBooleanCookieAttributes(requestData map[string]interface{}) (map[string]interface{}, error) {
	for _, flagName := range []string{"use_cookie", "cookie_secure", "cookie_same_site"} {
		if stringValue, ok := requestData[flagName]; ok {
			if _, ok = map[string]bool{"0": false, "1": true}[stringValue.(string)]; !ok {
				return nil, service.ErrInvalidRequest(fmt.Errorf("wrong value for %s (should have a boolean value (0 or 1))", flagName))
			}
			delete(requestData, flagName)
			if stringValue == "1" {
				requestData[flagName] = true
			}
		}
	}
	return requestData, nil
}

func extractOptionalParameter(query url.Values, paramName string, requestData map[string]interface{}) {
	if len(query[paramName]) != 0 {
		requestData[paramName] = query.Get(paramName)
	}
}

func createOrUpdateUser(userStore *database.UserStore, userData map[string]interface{}, domainConfig *domain.CtxConfig) int64 {
	var groupID int64
	err := userStore.WithExclusiveWriteLock().
		Where("login_id = ?", userData["login_id"]).PluckFirst("group_id", &groupID).Error()

	userData["latest_login_at"] = database.Now()
	userData["latest_activity_at"] = database.Now()
	userData["latest_profile_sync_at"] = database.Now()

	if defaultLanguage, ok := userData["default_language"]; ok && defaultLanguage == nil {
		userData["default_language"] = database.Default()
	}

	badges := userData["badges"]
	delete(userData, "badges")
	defer func() { userData["badges"] = badges }()

	if gorm.IsRecordNotFoundError(err) {
		selfGroupID := createGroupFromLogin(userStore.Groups(), userData["login"].(string), domainConfig)
		userData["temp_user"] = 0
		userData["registered_at"] = database.Now()
		userData["group_id"] = selfGroupID

		service.MustNotBeError(userStore.InsertMap(userData))
		service.MustNotBeError(userStore.Attempts().InsertMap(map[string]interface{}{
			"participant_id": selfGroupID,
			"id":             0,
			"creator_id":     selfGroupID,
			"created_at":     database.Now(),
		}))

		return selfGroupID
	}
	service.MustNotBeError(err)

	found, err := userStore.GroupGroups().WithExclusiveWriteLock().Where("parent_group_id = ?", domainConfig.NonTempUsersGroupID).
		Where("child_group_id = ?", groupID).HasRows()
	service.MustNotBeError(err)
	groupsGroupsToCreate := make([]map[string]interface{}, 0, 1)
	if !found {
		groupsGroupsToCreate = append(groupsGroupsToCreate,
			map[string]interface{}{"parent_group_id": domainConfig.NonTempUsersGroupID, "child_group_id": groupID})
	}

	service.MustNotBeError(userStore.GroupGroups().CreateRelationsWithoutChecking(groupsGroupsToCreate))
	delete(userData, "default_language")
	service.MustNotBeError(userStore.ByID(groupID).UpdateColumn(userData).Error())
	return groupID
}

func createGroupFromLogin(store *database.GroupStore, login string, domainConfig *domain.CtxConfig) (selfGroupID int64) {
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError("groups", func(retryIDStore *database.DataStore) error {
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
		{"parent_group_id": domainConfig.NonTempUsersGroupID, "child_group_id": selfGroupID},
	}))

	return selfGroupID
}
