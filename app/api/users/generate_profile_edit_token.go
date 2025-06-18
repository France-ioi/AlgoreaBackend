package users

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/encrypt"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const profileEditTokenLifetime = 30 * time.Minute

// swagger:model generateProfileEditTokenResponse
type generateProfileEditTokenResponse struct {
	// The ProfileEditToken encoded as hex.
	// required:true
	ProfileEditToken string `json:"token"`
	// The algorithm used to encrypt the token (for now, it can only be `AES-256-GCM`).
	// For `AES-256-GCM`, the `nonce` is the first 12 bytes of the token, and the `ciphertext` is the rest,
	// once the `token` is decoded from hex to binary.
	// required:true
	Alg string `json:"alg"`
	// This field is not really present, it is here only to document the content of the token.
	// required:false
	TokenForDoc *ProfileEditToken `json:"token_not_present_only_for_doc,omitempty"`
}

// ProfileEditToken permits a requester user to edit the profile of a target user.
// swagger:model ProfileEditToken
type ProfileEditToken struct {
	// `loginID` of the user who requested the token.
	// required:true
	RequesterID string `json:"requester_id"`
	// `loginID` of the user whose profile is to be edited.
	// required:true
	TargetID string `json:"target_id"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp int64 `json:"exp,string"`
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
//			* the `current user` must be a manager of a group to which the `target user` is a descendant, and
//			* this group must have `require_personal_info_access_approval` set to `edit`,
//			* both the `current user` and the `target user` must have a `login_id`,
//
//		otherwise, a forbidden error is returned.
//
//	parameters:
//		- name: target_user_id
//			in: path
//			type: integer
//			format: int64
//			description: The ID of the user whose profile is to be edited.
//			required: true
//
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
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) generateProfileEditToken(rw http.ResponseWriter, r *http.Request) error {
	targetUserID, err := service.ResolveURLQueryPathInt64Field(r, "target_user_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	var targetUserLoginID *int64
	if user.LoginID != nil {
		targetUserLoginID, err = getLoginIDForProfileEditing(store, user, targetUserID)
		service.MustNotBeError(err)
	}

	// Checks rights.
	if targetUserLoginID == nil {
		return service.ErrAPIInsufficientAccessRights
	}

	response := new(generateProfileEditTokenResponse)

	response.ProfileEditToken, response.Alg = srv.getProfileEditToken(*user.LoginID, *targetUserLoginID)

	render.Respond(rw, r, response)

	return nil
}

func (srv *Service) getProfileEditToken(requesterLoginID, targetLoginID int64) (token, algorithm string) {
	expirationTime := time.Now().Add(profileEditTokenLifetime)

	profileEditToken := ProfileEditToken{
		RequesterID: strconv.FormatInt(requesterLoginID, 10),
		TargetID:    strconv.FormatInt(targetLoginID, 10),
		Exp:         expirationTime.Unix(),
	}

	jsonToken, _ := json.Marshal(profileEditToken)

	key := []byte(srv.AuthConfig.GetString("clientSecret")[0:32])
	cipherText := encrypt.AES256GCM(key, jsonToken)

	hexCipher := hex.EncodeToString(cipherText)

	return hexCipher, "AES-256-GCM"
}

// getLoginIDForProfileEditing returns `login_id` of the given target user
// if the requesting user can edit the profile of the target user:
//  1. the requesting user needs to be a manager of a group to which the target user is a descendant, and
//  2. this group must have `require_personal_info_access_approval` set to `edit`, and
//  3. the target user must be a user,
//
// otherwise, it returns nil.
func getLoginIDForProfileEditing(s *database.DataStore, requestingUser *database.User, targetUserID int64) (*int64, error) {
	var targetUserLoginID *int64

	err := s.ActiveGroupAncestors().
		ManagedByUser(requestingUser).
		Joins(`JOIN groups_ancestors AS target_user_group_ancestor
									 ON target_user_group_ancestor.ancestor_group_id = groups_ancestors_active.child_group_id AND
											target_user_group_ancestor.child_group_id = ?`, targetUserID).
		Joins("JOIN `users` AS target_user ON target_user.group_id = target_user_group_ancestor.child_group_id").
		Joins("JOIN `groups` AS target_user_group ON target_user_group.id = target_user_group_ancestor.ancestor_group_id").
		Where("target_user_group.require_personal_info_access_approval = 'edit'").
		PluckFirst("target_user.login_id", &targetUserLoginID).
		Error()

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return targetUserLoginID, err
}
