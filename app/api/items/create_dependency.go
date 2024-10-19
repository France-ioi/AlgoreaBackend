package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model
type itemDependencyCreateRequest struct {
	// minimum: 0
	// maximum: 100
	// default: 100
	Score int32 `json:"score" validate:"min=0,max=100"`
	// required: true
	GrantContentView bool `json:"grant_content_view"`
}

// swagger:operation POST /items/{dependent_item_id}/prerequisites/{prerequisite_item_id} items itemDependencyCreate
//
//	---
//	summary: Create an item dependency
//	description: >
//
//		Creates an item dependency with parameters from the input data without any effect to access rights.
//
//		The user should have:
//			* `can_edit` >= 'all' on the `{dependent_item_id}` item,
//			* `can_view` >= 'info' on the `{prerequisite_item_id}` item,
//			* if `grant_content_view` = true, the user should also have `can_grant_view` >= 'content'
//				on the `{dependent_item_id}` item,
//				otherwise the "forbidden" response is returned.
//
//	parameters:
//		- name: dependent_item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: prerequisite_item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- in: body
//			name: data
//			required: true
//			description: The item dependency to create
//			schema:
//				"$ref": "#/definitions/itemDependencyCreateRequest"
//	responses:
//		"201":
//			description: Created. The request has successfully created the item dependency.
//			schema:
//				"$ref": "#/definitions/createdResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"422":
//			"$ref": "#/responses/unprocessableEntityResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createDependency(w http.ResponseWriter, r *http.Request) service.APIError {
	dependentItemID, err := service.ResolveURLQueryPathInt64Field(r, "dependent_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	prerequisiteItemID, err := service.ResolveURLQueryPathInt64Field(r, "prerequisite_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	input := itemDependencyCreateRequest{}
	formData := formdata.NewFormData(&input)
	err = formData.ParseJSONRequestData(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	apiError := service.NoError
	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		if !user.CanViewItemInfo(store.WithExclusiveWriteLock(), prerequisiteItemID) {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}

		permissionsQuery := store.Permissions().
			AggregatedPermissionsForItemsOnWhichGroupHasPermission(user.GroupID, "edit", "all").
			Where("item_id = ?", dependentItemID).
			WithExclusiveWriteLock()
		if input.GrantContentView {
			permissionsQuery = permissionsQuery.HavingMaxPermissionAtLeast("grant_view", "content")
		}

		var found bool
		found, err = permissionsQuery.HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}

		err = store.ItemDependencies().InsertMap(map[string]interface{}{
			"item_id":            prerequisiteItemID,
			"dependent_item_id":  dependentItemID,
			"score":              valueOrDefault(formData, "score", input.Score, database.Default()),
			"grant_content_view": input.GrantContentView,
		})
		if err != nil && database.IsDuplicateEntryError(err) {
			apiError = service.ErrUnprocessableEntity(errors.New("the dependency already exists"))
			return apiError.Error // rollback
		}

		return err
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess[*struct{}](nil)))

	return service.NoError
}
