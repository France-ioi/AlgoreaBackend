package items

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemWithDefaultLanguageIDAndOptionalType struct {
	item `json:"item,squash"`
	// new `default_language_id` of the item can only be set to a language
	// for that an `items_strings` row exists
	DefaultLanguageID int64 `json:"default_language_id" validate:"default_language_id"`
	// enum: Root,Category,Chapter,Task,Course
	Type string `json:"type" validate:"oneof=Root Category Chapter Task Course"`
}

// updateItemRequest is the expected input for item updating
// swagger:model itemEditRequest
type updateItemRequest struct {
	itemWithDefaultLanguageIDAndOptionalType `json:"item,squash"`
	Children                                 []itemChild `json:"children" validate:"children"`
}

func (in *updateItemRequest) checkItemsRelationsCycles(store *database.DataStore, itemID int64) bool {
	if len(in.Children) == 0 {
		return true
	}
	ids := make([]int64, len(in.Children)+1)
	for index := range in.Children {
		if in.Children[index].ItemID == itemID {
			return false
		}
		ids[index] = in.Children[index].ItemID
	}
	var count int64
	service.MustNotBeError(store.ItemAncestors().WithWriteLock().
		Where("child_item_id = ?", itemID).
		Where("ancestor_item_id IN (?)", ids).Count(&count).Error())
	return count == 0
}

// swagger:operation PUT /items/{item_id} items itemEdit
// ---
// summary: Edit an item
// description: >
//
//   Changes the given item's properties using the input data.
//
//
//   If the `children` array is given, the service removes relations between the item and its former children,
//   inserts relations between the item and its new children into `items_items`, and propagates `permissions_generated`.
//
//
//   If a `duration` is added and `contest_participants_group_id` is NULL, the service creates a participants group,
//   links `contest_participants_group_id` to it, and gives this group 'can_view:content' permission on the new item.
//
//
//   The user should have
//
//     * `can_edit` >= 'all' on the item, otherwise the "forbidden" response is returned;
//     * `can_view` != 'none' on the `children` items (if any), otherwise the "bad request"
//       response is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - in: body
//   name: data
//   required: true
//   description: New item property values
//   schema:
//     "$ref": "#/definitions/itemEditRequest"
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateItem(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	input := updateItemRequest{}
	formData := formdata.NewFormData(&input)

	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var childrenPermissions []permission
		registerChildrenValidator(formData, store, user, &childrenPermissions)
		registerDefaultLanguageIDValidator(formData, store, itemID)

		err = formData.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		var participantsGroupID *int64
		err = store.Permissions().MatchingUserAncestors(user).WithWriteLock().
			Joins("JOIN items ON items.id = item_id").
			Where("item_id = ?", itemID).
			WherePermissionIsAtLeast("edit", "all").
			PluckFirst("items.contest_participants_group_id", &participantsGroupID).Error()

		if gorm.IsRecordNotFoundError(err) {
			apiError = service.ErrForbidden(errors.New("no access rights to edit the item"))
			return apiError.Error // rollback
		}
		service.MustNotBeError(err)

		itemData := formData.ConstructPartialMapForDB("itemWithDefaultLanguageIDAndOptionalType")

		if itemData["duration"] != nil && participantsGroupID == nil {
			service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(s *database.DataStore) error {
				participantsGroupIDTmp := s.NewID()
				participantsGroupID = &participantsGroupIDTmp
				service.MustNotBeError(s.Groups().InsertMap(map[string]interface{}{
					"id": participantsGroupID, "type": "ContestParticipants",
					"name": fmt.Sprintf("%d-participants", itemID),
				}))
				return s.PermissionsGranted().InsertMap(map[string]interface{}{
					"group_id":       participantsGroupID,
					"item_id":        itemID,
					"giver_group_id": -1,
					"can_view":       "content",
				})
			}))
			itemData["contest_participants_group_id"] = participantsGroupID
		}

		service.MustNotBeError(store.Items().Where("id = ?", itemID).UpdateColumn(itemData).Error())
		if formData.IsSet("children") {
			err = store.WithNamedLock("items_items", 3*time.Second, func(lockedStore *database.DataStore) error {
				service.MustNotBeError(lockedStore.ItemItems().Delete("parent_item_id = ?", itemID).Error())

				if !input.checkItemsRelationsCycles(lockedStore, itemID) {
					apiError = service.ErrForbidden(errors.New("an item cannot become an ancestor of itself"))
					return apiError.Error // rollback
				}

				service.MustNotBeError(lockedStore.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
					parentChildSpec := constructItemsItemsForChildren(childrenPermissions, input.Children, retryStore, itemID)
					insertItemItems(retryStore, parentChildSpec)
					return nil
				}))
				return lockedStore.ItemItems().After()
			})
		}
		return err
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}

func registerDefaultLanguageIDValidator(formData *formdata.FormData, store *database.DataStore, itemID int64) {
	formData.RegisterValidation("default_language_id", constructDefaultLanguageIDValidator(formData, store, itemID))
	formData.RegisterTranslation("default_language_id",
		"default language should exist and there should be item's strings in this language")
}

// constructDefaultLanguageIDValidator constructs a validator for the DefaultLanguageID field.
// The validator checks that the language exists and there is an items_strings row in this language for the item.
func constructDefaultLanguageIDValidator(formData *formdata.FormData, store *database.DataStore, itemID int64) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		if !formData.IsValid("default_language_id") {
			return true
		}
		found, err := store.Languages().ByID(fl.Field().Interface().(int64)).WithWriteLock().
			Joins("JOIN items_strings ON items_strings.language_id = languages.id AND items_strings.item_id = ?", itemID).
			HasRows()
		service.MustNotBeError(err)
		return found
	})
}
