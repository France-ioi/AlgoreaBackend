package currentuser

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model
type userDataUpdateRequest struct {
	DefaultLanguage string `json:"default_language"`
}

// swagger:operation PUT /current-user users userDataUpdate
//
//	---
//	summary: Update user's data
//	description: Allows changing the user's default language
//	parameters:
//		- name: data
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/userDataUpdateRequest"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) update(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)

	var requestData userDataUpdateRequest
	formData := formdata.NewFormData(&requestData)
	err := formData.ParseJSONRequestData(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	// the user middleware has already checked that the user exists so we just ignore the case where nothing is updated
	service.MustNotBeError(srv.GetStore(httpRequest).Users().ByID(user.GroupID).UpdateColumn(requestData).Error())

	response := service.Response[*struct{}]{Success: true, Message: "updated"}
	render.Respond(responseWriter, httpRequest, &response)

	return nil
}
