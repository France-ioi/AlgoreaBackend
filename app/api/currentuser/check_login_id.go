package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model loginIDCheckData
type loginIDCheckData struct {
	// required: true
	LoginIDMatched bool `json:"login_id_matched"`
}

// swagger:operation GET /current-user/check-login-id users loginIDCheck
//
//	---
//	summary: Check if a login id is the current user's login id
//	description: Checks if a given `{login_id}` matches the one of the current user.
//	parameters:
//		- name: login_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//				description: OK. Success response with the result
//				schema:
//					"$ref": "#/definitions/loginIDCheckData"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) checkLoginID(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	loginID, err := service.ResolveURLQueryGetInt64Field(r, "login_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	render.Respond(w, r, &loginIDCheckData{
		LoginIDMatched: user.LoginID != nil && *user.LoginID == loginID,
	})
	return service.NoError
}
