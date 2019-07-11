package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
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

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		var login string
		var userID int64
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
			userID = retryIDStore.NewID()
			return retryIDStore.RetryOnDuplicateKeyError("sLogin", "login", func(retryLoginStore *database.DataStore) error {
				login = fmt.Sprintf("tmp-%d", rand.Int31n(99999999-10000000+1)+10000000)
				return retryLoginStore.Users().InsertMap(map[string]interface{}{
					"ID":                userID,
					"loginID":           0,
					"sLogin":            login,
					"tempUser":          true,
					"sRegistrationDate": gorm.Expr("NOW()"),
					"idGroupSelf":       nil,
					"idGroupOwned":      nil,
					"sLastIP":           strings.SplitN(r.RemoteAddr, ":", 2)[0],
				})
			})
		}))
		var selfGroupID int64
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *database.DataStore) error {
			selfGroupID = retryIDStore.NewID()
			return retryIDStore.Groups().InsertMap(map[string]interface{}{
				"ID":           selfGroupID,
				"sName":        login,
				"sType":        "UserSelf",
				"sDescription": login,
				"sDateCreated": gorm.Expr("NOW()"),
				"bOpened":      false,
				"bSendEmails":  false,
			})
		}))
		service.MustNotBeError(store.Users().ByID(userID).UpdateColumn("idGroupSelf", selfGroupID).Error())

		var rootTempGroupID int64
		service.MustNotBeError(store.Groups().
			Where("sType = 'UserSelf'").
			Where("sName = 'RootTemp'").
			Where("sTextId = 'RootTemp'").PluckFirst("ID", &rootTempGroupID).Error())
		service.MustNotBeError(store.GroupGroups().CreateRelation(rootTempGroupID, selfGroupID))

		var err error
		token, expiresIn, err = store.Sessions().CreateNewTempSession(userID)
		return err
	}))

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"access_token": token,
		"expires_in":   expiresIn,
	})))
	return service.NoError
}
