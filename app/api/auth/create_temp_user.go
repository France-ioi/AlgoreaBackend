package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	authlib "github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /auth/temp-user users auth userCreateTmp
// ---
// summary: Create a temporary user
// description: Creates a temporary user and generates an access token valid for 2 hours
//
//   * No “Authorization” header should be present
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
	if len(r.Header["Authorization"]) != 0 {
		return service.ErrInvalidRequest(errors.New("'Authorization' header should not be present"))
	}

	var token string
	var expiresIn int32
	var userID int64

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		var login string
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
			userID = retryIDStore.NewID()
			return retryIDStore.RetryOnDuplicateKeyError("login", "login", func(retryLoginStore *database.DataStore) error {
				login = fmt.Sprintf("tmp-%d", rand.Int31n(99999999-10000000+1)+10000000)
				return retryLoginStore.Users().InsertMap(map[string]interface{}{
					"id":             userID,
					"login_id":       0,
					"login":          login,
					"temp_user":      true,
					"registered_at":  database.Now(),
					"self_group_id":  nil,
					"owned_group_id": nil,
					"last_ip":        strings.SplitN(r.RemoteAddr, ":", 2)[0],
				})
			})
		}))
		var selfGroupID int64
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
			selfGroupID = retryIDStore.NewID()
			return retryIDStore.Groups().InsertMap(map[string]interface{}{
				"id":          selfGroupID,
				"name":        login,
				"type":        "UserSelf",
				"description": login,
				"created_at":  database.Now(),
				"opened":      false,
				"send_emails": false,
			})
		}))
		service.MustNotBeError(store.Users().ByID(userID).UpdateColumn("self_group_id", selfGroupID).Error())

		domainConfig := domain.ConfigFromContext(r.Context())
		service.MustNotBeError(store.GroupGroups().CreateRelationsWithoutChecking(
			[]database.ParentChild{{ParentID: domainConfig.RootTempGroupID, ChildID: selfGroupID}}))

		var err error
		token, expiresIn, err = authlib.CreateNewTempSession(store.Sessions(), userID)
		return err
	}))

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": token,
		"expires_in":   expiresIn,
	})))
	return service.NoError
}
