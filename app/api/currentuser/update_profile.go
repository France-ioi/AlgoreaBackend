package currentuser

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loginmodule"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Without a deadline, a hung login module would hold the request until the client gives up;
// this one stays well below the server's default write timeout (60s) so that the 502 still reaches the client.
const profileUpdateLoginModuleTimeout = 15 * time.Second

// swagger:model
type userProfileUpdateRequest struct {
	// The profile attributes to set, forwarded unchanged to the login module, which decides
	// which keys and values it accepts.
	// required: true
	Profile map[string]string `json:"profile" validate:"set,required"`
}

// swagger:operation PUT /current-user/profile users userProfileUpdate
//
//	---
//	summary: Update the current user's profile
//	description:
//		Forwards the profile unchanged to the login module (`PUT /user_api/account`) with the caller's own token,
//		then updates the local user info cache stored in the `users` table and badges from the profile
//		the login module returns, as `PUT /current-user/refresh` does.
//
//		The service is enabled only when the `auth.profileUpdateEnabled` setting is true.
//		Otherwise, the not found error is returned.
//
//		* If the current user is temporary, the forbidden error is returned.
//
//		* If the login module refuses the profile (4xx), the bad request error is returned
//			with the login module's message as `error_text`.
//
//		* If the login module fails otherwise or does not answer in time, the bad gateway error is returned.
//	parameters:
//		- name: data
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/userProfileUpdateRequest"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
//		"502":
//			"$ref": "#/responses/badGatewayResponse"
func (srv *Service) updateProfile(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	// Off by default: login modules that do not implement PUT /user_api/account must not expose the route.
	if !srv.AuthConfig.GetBool("profileUpdateEnabled") {
		return service.ErrNotFound(nil)
	}

	user := srv.GetUser(httpRequest)
	if user.IsTempUser {
		return service.ErrAPIInsufficientAccessRights
	}

	var requestData userProfileUpdateRequest
	formData := formdata.NewFormData(&requestData)
	if err := formData.ParseJSONRequestData(httpRequest); err != nil {
		return service.ErrInvalidRequest(err)
	}

	accessToken := auth.BearerTokenFromContext(httpRequest.Context())
	loginModuleContext, cancel := context.WithTimeout(httpRequest.Context(), profileUpdateLoginModuleTimeout)
	defer cancel()
	userProfile, err := loginmodule.NewClient(srv.AuthConfig.GetString("loginModuleURL")).
		UpdateUserProfile(loginModuleContext, accessToken, requestData.Profile)
	if err != nil {
		var rejectedError *loginmodule.ProfileRejectedError
		if errors.As(err, &rejectedError) {
			return service.ErrInvalidRequest(rejectedError)
		}
		logging.GetLogEntry(httpRequest).Warnf("The login module failed to update the profile of user %d: %s", user.GroupID, err)
		// Any other failure (including a timeout) is the login module's, not the caller's:
		// it must not be reported as a 408 or a 500.
		return service.ErrBadGateway(errors.New("the login module failed to update the profile"))
	}

	srv.storeUserProfile(httpRequest, user.GroupID, userProfile)

	response := service.UpdateSuccess[*struct{}](nil)
	render.Respond(responseWriter, httpRequest, &response)

	return nil
}
