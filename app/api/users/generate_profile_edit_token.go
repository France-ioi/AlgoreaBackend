package users

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model generateProfileEditTokenResponse
type generateProfileEditTokenResponse struct {
	// The ProfileEditToken
	// required:true
	ProfileEditToken string `json:"token"`
	// The algorithm used to encrypt the token
	// required:true
	Alg string `json:"alg"`
	// This field is not really present, it is here only to document the content of the token.
	// required:false
	TokenForDoc *ProfileEditToken `json:"token_not_present_only_for_doc,omitempty"`
}

// ProfileEditToken permits a requester user to edit the profile of a target user.
// swagger:model ProfileEditToken
type ProfileEditToken struct {
	// User who requested the token.
	// required:true
	RequesterID string `json:"requester_id"`
	// User whose profile is to be edited.
	// required:true
	TargetID string `json:"target_id"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp string `json:"exp"`
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

	response.ProfileEditToken, response.Alg = srv.getProfileEditToken(user.GroupID, targetUserID)

	render.Respond(rw, r, response)

	return service.NoError
}

func (srv *Service) getProfileEditToken(requesterID, targetID int64) (token string, algorithm string) {
	thirtyMinutesLater := time.Now().Add(time.Minute * 30)

	profileEditToken := ProfileEditToken{
		RequesterID: strconv.FormatInt(requesterID, 10),
		TargetID:    strconv.FormatInt(targetID, 10),
		Exp:         strconv.FormatInt(thirtyMinutesLater.Unix(), 10),
	}

	jsonToken, err := json.Marshal(profileEditToken)
	service.MustNotBeError(err)

	key := []byte(srv.AuthConfig.GetString("clientSecret")[0:32])
	block, err := aes.NewCipher(key)
	service.MustNotBeError(err)

	gcm, err := cipher.NewGCM(block)
	service.MustNotBeError(err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	service.MustNotBeError(err)

	cipherText := gcm.Seal(nonce, nonce, jsonToken, nil)

	hexCipher := hex.EncodeToString(cipherText)

	return hexCipher, "AES-256-GCM"
}