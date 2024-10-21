package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /items/{dependent_item_id}/prerequisites/{prerequisite_item_id} items itemDependencyDelete
//
//	---
//	summary: Delete a specific item-dependency rule
//	description: Deletes the rule without any effect to access rights.
//
//
//						 * The current-user must have `can_edit` = 'all' on the `{dependent_item_id}`,
//							 otherwise the 'forbidden' error is returned.
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
//	responses:
//		"200":
//			"$ref": "#/responses/deletedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) deleteDependency(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	dependentItemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "dependent_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	prerequisiteItemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "prerequisite_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)

	apiError := service.NoError
	err = srv.GetStore(httpReq).InTransaction(func(store *database.DataStore) error {
		var found bool
		found, err = store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("edit", "all").
			Where("item_id = ?", dependentItemID).WithExclusiveWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}
		return store.ItemDependencies().
			Delete("item_id = ? AND dependent_item_id = ?", prerequisiteItemID, dependentItemID).Error()
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(rw, httpReq, service.DeletionSuccess[*struct{}](nil)))
	return service.NoError
}
