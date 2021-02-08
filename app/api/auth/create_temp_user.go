package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /auth/temp-user auth tempUserCreate
// ---
// summary: Create a temporary user
// description: Creates a temporary user and generates an access token valid for 2 hours
//
//   * None of “Authorization” header and "access_token" cookie should be present
// parameters:
// - name: use_cookie
//   in: query
//   description: If 1, set a cookie instead of returning the OAuth2 code in the data
//   type: integer
//   enum: [0,1]
//   default: 0
// - name: cookie_secure
//   in: query
//   description: If 1, set the cookie with the `Secure` attribute
//   type: integer
//   enum: [0,1]
//   default: 0
// - name: cookie_same_site
//   in: query
//   description: If 1, set the cookie with the `SameSite`='Strict' attribute value and with `SameSite`='None' otherwise
//   type: integer
//   enum: [0,1]
//   default: 0
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
func (srv *Service) createTempUser(w http.ResponseWriter, r *http.Request) service.APIError {
	_, cookieErr := r.Cookie("access_token")
	if len(r.Header["Authorization"]) != 0 || cookieErr == nil {
		return service.ErrInvalidRequest(errors.New("neither 'Authorization' header nor 'access_token' cookie should not be present"))
	}

	cookieAttributes, apiError := srv.resolveCookieAttributesFromRequest(r)
	if apiError != service.NoError {
		return apiError
	}

	var token string
	var expiresIn int32

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		var login string
		var userID int64
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
			userID = retryIDStore.NewID()
			return retryIDStore.Groups().InsertMap(map[string]interface{}{
				"id":          userID,
				"type":        "User",
				"created_at":  database.Now(),
				"is_open":     false,
				"send_emails": false,
			})
		}))
		service.MustNotBeError(store.RetryOnDuplicateKeyError("login", "login", func(retryLoginStore *database.DataStore) error {
			login = fmt.Sprintf("tmp-%d", rand.Int31n(99999999-10000000+1)+10000000)
			return retryLoginStore.Users().InsertMap(map[string]interface{}{
				"login_id":      0,
				"login":         login,
				"temp_user":     true,
				"registered_at": database.Now(),
				"group_id":      userID,
				"last_ip":       strings.SplitN(r.RemoteAddr, ":", 2)[0],
			})
		}))

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

		domainConfig := domain.ConfigFromContext(r.Context())
		service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking(
			[]map[string]interface{}{{"parent_group_id": domainConfig.TempUsersGroupID, "child_group_id": userID}}))

		var err error
		token, expiresIn, err = auth.CreateNewTempSession(store.Sessions(), userID, cookieAttributes)
		return err
	}))

	srv.respondWithNewAccessToken(r, w, service.CreationSuccess,
		token, time.Now().Add(time.Duration(expiresIn)*time.Second), cookieAttributes)
	return service.NoError
}

func (srv *Service) resolveCookieAttributesFromRequest(r *http.Request) (*database.SessionCookieAttributes, service.APIError) {
	requestData, apiError := parseCookieAttributesForCreateTempUser(r)
	if apiError != service.NoError {
		return nil, apiError
	}
	cookieAttributes, apiError := srv.resolveCookieAttributes(r, requestData)
	if apiError != service.NoError {
		return nil, apiError
	}
	return cookieAttributes, service.NoError
}

func parseCookieAttributesForCreateTempUser(r *http.Request) (map[string]interface{}, service.APIError) {
	allowedParameters := []string{"use_cookie", "cookie_secure", "cookie_same_site"}
	requestData := make(map[string]interface{}, 2)
	query := r.URL.Query()
	for _, parameterName := range allowedParameters {
		extractOptionalParameter(query, parameterName, requestData)
	}

	return preprocessBooleanCookieAttributes(requestData)
}
