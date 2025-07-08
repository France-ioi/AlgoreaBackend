package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /auth/temp-user auth tempUserCreate
//
//	---
//	summary: Create a temporary user
//	description: Creates a temporary user and generates an access token valid for 2 hours.
//
//		If attributes of the old and the new 'access_token' cookies are different (or the token is returned in the JSON),
//		the old cookie gets deleted (otherwise, just overwritten).
//
//		* The "Authorization" header must not be given.
//
//		* When `{use_cookie}`=1, at least one of `{cookie_secure}` and `{cookie_same_site}` must be true.
//	parameters:
//		- name: default_language
//			in: query
//			type: string
//			maxLength: 3
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
//	responses:
//		"201":
//			description: "Created. Success response with the new access token"
//			in: body
//			schema:
//				"$ref": "#/definitions/userCreateTmpResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createTempUser(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	cookieAttributes, err := srv.resolveCookieAttributesFromRequest(httpRequest)
	service.MustNotBeError(err)

	if len(httpRequest.Header["Authorization"]) != 0 {
		return service.ErrInvalidRequest(errors.New("the 'Authorization' header must not be present"))
	}

	defaultLanguage := database.Default()
	if len(httpRequest.URL.Query()["default_language"]) != 0 {
		defaultLanguage = httpRequest.URL.Query().Get("default_language")
		const maxLanguageLength = 3
		if utf8.RuneCountInString(defaultLanguage.(string)) > maxLanguageLength {
			return service.ErrInvalidRequest(errors.New("the length of default_language should be no more than 3 characters"))
		}
	}

	var token string
	var expiresIn int32

	service.MustNotBeError(srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		userID := createTempUserGroup(store)
		login := createTempUser(store, userID, defaultLanguage,
			strings.SplitN(httpRequest.RemoteAddr, ":", 2)[0]) //nolint:mnd // cut off the port

		service.MustNotBeError(store.Groups().ByID(userID).UpdateColumn(map[string]interface{}{
			"name":        login,
			"description": login,
		}).Error())

		service.MustNotBeError(store.Attempts().InsertMap(map[string]interface{}{
			"participant_id": userID,
			"id":             0,
			"creator_id":     userID,
			"created_at":     database.Now(),
		}))

		domainConfig := domain.ConfigFromContext(httpRequest.Context())
		service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking(
			[]map[string]interface{}{{"parent_group_id": domainConfig.TempUsersGroupID, "child_group_id": userID}}))

		var err error
		token, expiresIn, err = auth.CreateNewTempSession(store, userID)
		return err
	}))

	srv.respondWithNewAccessToken(
		responseWriter, httpRequest, service.CreationSuccess[map[string]interface{}], token, expiresIn, cookieAttributes)
	return nil
}

func createTempUser(store *database.DataStore, userID int64, defaultLanguage interface{}, lastIP string) string {
	var login string
	service.MustNotBeError(store.RetryOnDuplicateKeyError("users", "login", "login", func(retryLoginStore *database.DataStore) error {
		const minLogin = int32(10000000)
		const maxLogin = int32(99999999)
		login = fmt.Sprintf("tmp-%d", rand.Int31n(maxLogin-minLogin+1)+minLogin)
		return retryLoginStore.Users().InsertMap(map[string]interface{}{
			"login_id":         0,
			"login":            login,
			"temp_user":        true,
			"registered_at":    database.Now(),
			"group_id":         userID,
			"default_language": defaultLanguage,
			"last_ip":          lastIP,
		})
	}))
	return login
}

func createTempUserGroup(store *database.DataStore) int64 {
	var userID int64
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError("groups", func(retryIDStore *database.DataStore) error {
		userID = retryIDStore.NewID()
		return retryIDStore.Groups().InsertMap(map[string]interface{}{
			"id":          userID,
			"type":        "User",
			"created_at":  database.Now(),
			"is_open":     false,
			"send_emails": false,
		})
	}))
	return userID
}

func (srv *Service) resolveCookieAttributesFromRequest(httpRequest *http.Request) (*auth.SessionCookieAttributes, error) {
	requestData, err := parseCookieAttributesForCreateTempUser(httpRequest)
	if err != nil {
		return nil, err
	}

	cookieAttributes, err := srv.resolveCookieAttributes(httpRequest, requestData)
	if err != nil {
		return nil, err
	}

	return cookieAttributes, nil
}

func parseCookieAttributesForCreateTempUser(r *http.Request) (map[string]interface{}, error) {
	allowedParameters := []string{"use_cookie", "cookie_secure", "cookie_same_site"}
	requestData := make(map[string]interface{}, len(allowedParameters))
	query := r.URL.Query()
	for _, parameterName := range allowedParameters {
		extractOptionalParameter(query, parameterName, requestData)
	}

	return preprocessBooleanCookieAttributes(requestData)
}
