package items

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const permissionsTokenLifetime = 2 * time.Hour

// swagger:operation POST /items/{item_id}/permissions-token items itemPermissionsTokenGenerate
//
//	---
//	summary: Generate a permissions token
//	description: >
//		Generates a signed JWS token that asserts the current user's permissions on an item.
//		The token contains the user's ID, item ID, and all permission fields
//		(can_view, can_grant_view, can_watch, can_edit, is_owner).
//
//
//		Restrictions:
//
//		* the user must have at least `can_view` >= 'info' on the item,
//
//		otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"201":
//			"$ref": "#/responses/permissionsTokenResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) generatePermissionsToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	var rawPermissions database.RawGeneratedPermissionFields
	err = store.Permissions().
		AggregatedPermissionsForItemsOnWhichGroupHasPermission(user.GroupID, "view", "info").
		Where("permissions.item_id = ?", itemID).
		Take(&rawPermissions).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	itemPermissions := rawPermissions.AsItemPermissions(store.PermissionsGranted())

	expiresIn := int32(permissionsTokenLifetime / time.Second)
	expirationTime := time.Now().Add(permissionsTokenLifetime)

	permissionsToken, err := (&token.Token[payloads.PermissionsToken]{Payload: payloads.PermissionsToken{
		UserID:       strconv.FormatInt(user.GroupID, 10),
		ItemID:       strconv.FormatInt(itemID, 10),
		CanView:      itemPermissions.CanView,
		CanGrantView: itemPermissions.CanGrantView,
		CanWatch:     itemPermissions.CanWatch,
		CanEdit:      itemPermissions.CanEdit,
		IsOwner:      itemPermissions.IsOwner,
		Exp:          expirationTime.Unix(),
	}}).Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.CreationSuccess(map[string]interface{}{
		"permissions_token": permissionsToken,
		"expires_in":        expiresIn,
	})))

	return nil
}
