package users

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// swagger:model generateProfileEditTokenResponse
type generateProfileEditTokenResponse struct {
	// The ProfileEditToken
	// required:true
	ProfileEditToken string `json:"token"`
	// This field is not really present, it is here only to document the content of the token.
	// required:false
	TokenForDoc *payloads.ProfileEditToken `json:"token_not_present_only_for_doc,omitempty"`
}

// swagger:operation POST /users/{target_user_id}/generate-profile-edit-token users generateProfileEditToken
//
//	---
//	summary: Get a token to edit the profile of another user
//	description: >
//		Gets a token to edit the profile of another user.
//
//
//		Restrictions:
//			* the current user must be a manager of a group of which the target user is a member, and
//			* the group must have `require_personal_info_access_approval` set to `edit`.
//		Otherwise, a forbidden error is returned.
//	responses:
//		"200":
//			description: OK. Success response with the token.
//			schema:
//				"$ref": "#/definitions/generateProfileEditTokenResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) generateProfileEditToken(rw http.ResponseWriter, r *http.Request) service.APIError {
	targetUserID, err := service.ResolveURLQueryPathInt64Field(r, "target_user_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	// Checks rights.
	if !user.CanEditProfile(store, targetUserID) {
		return service.InsufficientAccessRightsError
	}

	response := new(generateProfileEditTokenResponse)

	response.ProfileEditToken, err = srv.getProfileEditToken(user.GroupID, targetUserID)
	service.MustNotBeError(err)

	render.Respond(rw, r, response)

	return service.NoError
}

func (srv *Service) getProfileEditToken(requesterID, targetID int64) (string, error) {
	twoHoursLater := time.Now().Add(time.Hour * 2)

	profileEditToken, err := (&token.ProfileEdit{
		RequesterID: strconv.FormatInt(requesterID, 10),
		TargetID:    strconv.FormatInt(targetID, 10),
		Exp:         strconv.FormatInt(twoHoursLater.Unix(), 10),
	}).Sign(srv.TokenConfig.PrivateKey)

	return profileEditToken, err
}
