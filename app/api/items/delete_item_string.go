package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation DELETE /items/{item_id}/strings/{language_tag} items itemStringDelete
//
//	---
//	summary: Delete an item string entry
//	description: >
//
//		Deletes the corresponding `items_strings` row identified by `item_id` and `language_tag`.
//
//
//		`items_strings` having `language_tag` equal to the default language of the `item_id` item cannot be deleted
//	 	(the "unprocessable entity" error is returned in this case).
//
//
//		If the `items_strings` row does not exist, the "not found" error is returned.
//
//
//		The user should have `can_view` >= 'content' and `can_edit` >= 'all' on the item, otherwise the "forbidden" response is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: language_tag
//			in: path
//			type: string
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
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"422":
//			"$ref": "#/responses/unprocessableEntityResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) deleteItemString(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error
	user := srv.GetUser(httpRequest)

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	languageTag := chi.URLParam(httpRequest, "language_tag")

	err = srv.GetStore(httpRequest).InTransaction(func(store *database.DataStore) error {
		var found bool
		found, err = store.Permissions().MatchingUserAncestors(user).WithSharedWriteLock().
			Where("item_id = ?", itemID).
			WherePermissionIsAtLeast("view", "content").
			WherePermissionIsAtLeast("edit", "all").
			HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("no access rights to edit the item")) // rollback
		}

		result := store.ItemStrings().Delete("item_id = ? AND language_tag = ?", itemID, languageTag)
		if database.IsForeignKeyConstraintFailedOnDeletingOrUpdatingParentRowError(result.Error()) {
			return service.ErrUnprocessableEntity(
				errors.New("the item string cannot be deleted because its language is the default language of the item"))
		}
		service.MustNotBeError(result.Error())

		if result.RowsAffected() == 0 {
			return service.ErrNotFound(errors.New("no such item string"))
		}
		return nil // commit
	})

	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.DeletionSuccess[*struct{}](nil)))
	return nil
}
