package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model userData
type getInfoData struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	TempUser bool `json:"temp_user"`
	// required: true
	Login string `json:"login"`
	// The login-module profile JSON stored on the user. Not a projection of flat users columns.
	// required: true
	Profile *database.JSON `json:"profile"`
	// required: true
	DefaultLanguage string `json:"default_language"`
}

// swagger:operation GET /current-user users userData
//
//	---
//	summary: Get profile info for the current user
//	description: >
//		Returns `group_id`, `login`, `profile`, `default_language`, and `temp_user` for the current user.
//		`profile` is the login-module profile JSON stored on the user.
//		Retired flat personal attributes are not returned on this endpoint.
//	responses:
//		"200":
//				description: OK. Success response with user's data
//				schema:
//					"$ref": "#/definitions/userData"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getInfo(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)

	var userInfo getInfoData
	err := srv.GetStore(httpRequest).Users().ByID(user.GroupID).
		Select(`group_id, temp_user, login, default_language, profile`).
		Scan(&userInfo).Error()

	// This is very unlikely since the user middleware has already checked that the user exists
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, &userInfo)
	return nil
}
