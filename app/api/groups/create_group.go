package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model createGroupRequest
type createGroupRequest struct {
	// required: true
	// minLength: 1
	Name *string `json:"name" validate:"required,min=1"`
	// required: true
	// enum: Class,Team,Club,Friends,Other
	Type *string `json:"type" validate:"required,oneof=Class Team Club Friends Other"`
	// only if `type` = "Team"
	// required: true
	// type: string
	// format: int64
	// minimum: 1
	ItemID *int64 `json:"item_id,string" validate:"min=1"`
}

// swagger:operation POST /groups groups groupCreate
// ---
// summary: Create a group
// description: >
//
//   Creates a group with the input `name`, `type`, `sDateCreated` = now(), and default values in other columns.
//   If `item_id` is given:
//
//     * If `type` != "Team", returns the "badRequest" response
//
//     * Otherwise, checks that the authenticated user
//
//       * has grayed, partial or full access to the item (otherwise returns the "forbidden" response)
//
//       * sets this `item_id` as `idTeamItem` of the new group.
//
//   Also, the service sets the authenticated user as an owner of the group (with `sRole` = "owner").
//   After everything, it propagates group ancestors.
//
//
//   The user should not be temporary, otherwise the "forbidden" response is returned.
// consumes:
// - application/json
// parameters:
// - in: body
//   name: data
//   required: true
//   description: The group to create
//   schema:
//     "$ref": "#/definitions/createGroupRequest"
// responses:
//   "201":
//     description: "Created. Success response with the created group's ID"
//     schema:
//       type: object
//       required: [success, message, data]
//       properties:
//         success:
//           description: "true"
//           type: boolean
//           enum: [true]
//         message:
//           description: created
//           type: string
//           enum: [created]
//         data:
//           type: object
//           required: [id]
//           properties:
//             id:
//               type: string
//               format: int64
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	input := createGroupRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if input.ItemID != nil && *input.Type != "Team" {
		return service.ErrInvalidRequest(errors.New("only teams can be created with item_id set"))
	}

	if user.SelfGroupID == nil || user.OwnedGroupID == nil {
		return service.InsufficientAccessRightsError
	}

	apiError := service.NoError
	var groupID int64
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		if input.ItemID != nil {
			hasRows, itemErr := store.Raw("SELECT 1 FROM ? AS access_rights",
				store.GroupItems().AccessRightsForItemsVisibleToUser(user).Where("idItem = ?", *input.ItemID).
					WithWriteLock().SubQuery()).HasRows()
			service.MustNotBeError(itemErr)
			if !hasRows {
				apiError = service.InsufficientAccessRightsError
				return apiError.Error // rollback
			}
		}
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			groupID = retryStore.NewID()
			return retryStore.Groups().InsertMap(map[string]interface{}{
				"ID":           groupID,
				"sName":        input.Name,
				"sType":        input.Type,
				"idTeamItem":   input.ItemID,
				"sDateCreated": database.Now(),
			})
		}))
		return store.GroupGroups().CreateRelationsWithoutChecking([]database.ParentChild{
			{ParentID: *user.OwnedGroupID, ChildID: groupID, Role: "owner"},
		})
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	response := struct {
		GroupID int64 `json:"id,string"`
	}{GroupID: groupID}
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(&response)))
	return service.NoError
}
