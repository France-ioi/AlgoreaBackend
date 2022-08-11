package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /items/{item_id} items itemDelete
// ---
// summary: Delete an item
// description: >
//   Removes an item and objects linked to it.
//
//
//   The service deletes `answers`, `groups_contest_items`,
//   `item_dependencies` (by `item_id` and `dependent_item_id`),
//   `items_ancestors` (by `child_item_id`), `items_items` (by `child_item_id`), `items_strings`,
//   `permissions_generated`, `permissions_granted`, `permissions_propagate`, `results`
//   linked to the item.
//
//
//   The authenticated user should be an owner of the `{item_id}`, otherwise the "forbidden" error is returned.
//
//   Also, the item must not have any children, otherwise the "unprocessable entity" error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/deletedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) deleteItem(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	apiErr := service.NoError

	err = srv.GetStore(r).InTransaction(func(s *database.DataStore) error {
		var found bool
		found, err = s.Permissions().MatchingUserAncestors(user).Where("item_id = ?", itemID).
			Where("is_owner_generated").WithWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		found, err = s.ItemItems().ChildrenOf(itemID).WithWriteLock().HasRows()
		service.MustNotBeError(err)
		if found {
			apiErr = service.ErrUnprocessableEntity(errors.New("the item must not have children"))
			return apiErr.Error // rollback
		}

		return s.Items().DeleteItem(itemID)
	})

	if apiErr != service.NoError {
		return apiErr
	}

	service.MustNotBeError(err)
	service.MustNotBeError(render.Render(w, r, service.DeletionSuccess(nil)))
	return service.NoError
}
