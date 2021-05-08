package users

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// UserViewResponsePersonalInfo contains first_name and last_name
type UserViewResponsePersonalInfo struct {
	// Nullable
	FirstName *string `json:"first_name"`
	// Nullable
	LastName *string `json:"last_name"`
}

// swagger:model
type userViewResponse struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	TempUser bool `json:"temp_user"`
	// required: true
	Login string `json:"login"`
	// Nullable
	// required: true
	FreeText *string `json:"free_text"`
	// Nullable
	// required: true
	WebSite *string `json:"web_site"`

	*UserViewResponsePersonalInfo

	ShowPersonalInfo bool `json:"-"`
}

// swagger:operation GET /users/{user_id} users userView
// ---
// summary: Get profile info for a user
// description: Returns data from the `users` table for the given `{user_id}`
//              (`first_name` and `last_name` are only shown for the authenticated user or
//               if the user approved access to their personal info for some group
//               managed by the authenticated user).
// parameters:
// - name: user_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// responses:
//   "200":
//     description: OK. Success response with user's data
//     schema:
//       "$ref": "#/definitions/userViewResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUser(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	userID, err := service.ResolveURLQueryPathInt64Field(r, "user_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var userInfo userViewResponse
	err = srv.Store.Users().ByID(userID).
		Select(`
			group_id, temp_user, login, free_text, web_site,
			users.group_id = ? OR personal_info_view_approvals.approved AS show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS last_name`,
			user.GroupID, user.GroupID, user.GroupID).
		WithPersonalInfoViewApprovals(user).
		Scan(&userInfo).Error()

	if err == gorm.ErrRecordNotFound {
		return service.ErrNotFound(errors.New("no such user"))
	}
	service.MustNotBeError(err)

	if !userInfo.ShowPersonalInfo {
		userInfo.UserViewResponsePersonalInfo = nil
	}

	render.Respond(w, r, &userInfo)
	return service.NoError
}
