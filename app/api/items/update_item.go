package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemWithDefaultLanguageTagAndOptionalType struct {
	item `json:"item,squash"`
	// new `default_language_tag` of the item can only be set to a language
	// for that an `items_strings` row exists
	// minLength: 1
	// maxLength: 6
	DefaultLanguageTag string `json:"default_language_tag" validate:"min=1,max=6"`
}

// updateItemRequest is the expected input for item updating
// swagger:model itemEditRequest
type updateItemRequest struct {
	itemWithDefaultLanguageTagAndOptionalType `json:"item,squash"`
	Children                                  []itemChild `json:"children" validate:"children,children_allowed,dive,child_type_non_skill"`
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
//   (The only allowed parent-child relations are skills-*, chapter-task, chapter-course, chapter-chapter.
//   Otherwise the "bad request" error is returned.)
//
//
//   If a `duration` is added and `contest_participants_group_id` is NULL, the service creates a participants group,
//   links `contest_participants_group_id` to it, and gives this group 'can_view:content' permission on the new item.
//
//
//   The user should have
//
//     * `can_edit` >= 'children' on the item to edit children or `can_edit` >= 'all' to edit the item's properties,
//       otherwise the "forbidden" response is returned;
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
		var itemInfo struct {
			ContestParticipantsGroupID *int64
			Type                       string
			CanEditGenerated           string
		}
		err = store.Permissions().MatchingUserAncestors(user).WithWriteLock().
			Joins("JOIN items ON items.id = item_id").
			Where("item_id = ?", itemID).
			HavingMaxPermissionAtLeast("edit", "children").
			Select("items.contest_participants_group_id, items.type, MAX(can_edit_generated) AS can_edit_generated").
			Group("item_id").
			Scan(&itemInfo).Error()

		if gorm.IsRecordNotFoundError(err) {
			apiError = service.ErrForbidden(errors.New("no access rights to edit the item"))
			return apiError.Error // rollback
		}
		service.MustNotBeError(err)

		var childrenInfoMap map[int64]permissionAndType
		registerChildrenValidator(formData, store, user, itemInfo.Type, &childrenInfoMap)
		formData.RegisterValidation("child_type_non_skill", constructUpdateItemChildTypeNonSkillValidator(itemInfo.Type, &childrenInfoMap))
		formData.RegisterTranslation("child_type_non_skill", "a skill cannot be a child of a non-skill item")

		err = formData.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		itemData := formData.ConstructPartialMapForDB("itemWithDefaultLanguageTagAndOptionalType")
		if len(itemData) == 0 && !formData.IsSet("children") {
			return nil // Nothing to do
		}

		if len(itemData) > 0 &&
			store.PermissionsGranted().PermissionIndexByKindAndName("edit", itemInfo.CanEditGenerated) <
				store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all") {
			apiError = service.ErrForbidden(errors.New("no access rights to edit the item's properties"))
			return apiError.Error // rollback
		}

		apiError = updateItemInDB(itemData, itemInfo.ContestParticipantsGroupID, store, itemID)
		if apiError != service.NoError {
			return apiError.Error // rollback
		}

		apiError, err = updateChildrenAndRunListeners(formData, store, itemID, &input, childrenInfoMap)
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

func updateItemInDB(itemData map[string]interface{}, participantsGroupID *int64, store *database.DataStore, itemID int64) service.APIError {
	if itemData["duration"] != nil && participantsGroupID == nil {
		createdParticipantsGroupID := createContestParticipantsGroup(store, itemID)
		itemData["contest_participants_group_id"] = createdParticipantsGroupID
	}

	err := store.Items().Where("id = ?", itemID).UpdateColumn(itemData).Error()
	// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
	// (no items_strings for the new default_language_tag)
	if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1452 {
		return service.ErrInvalidRequest(formdata.FieldErrors{"default_language_tag": []string{
			"default language should exist and there should be item's strings in this language",
		}})
	}
	service.MustNotBeError(err)
	return service.NoError
}

func updateChildrenAndRunListeners(formData *formdata.FormData, store *database.DataStore, itemID int64,
	input *updateItemRequest, childrenPermissionMap map[int64]permissionAndType) (apiError service.APIError, err error) {
	if formData.IsSet("children") {
		err = store.WithNamedLock("items_items", 3*time.Second, func(lockedStore *database.DataStore) error {
			service.MustNotBeError(lockedStore.ItemItems().Delete("parent_item_id = ?", itemID).Error())

			if !input.checkItemsRelationsCycles(lockedStore, itemID) {
				apiError = service.ErrForbidden(errors.New("an item cannot become an ancestor of itself"))
				return apiError.Error // rollback
			}

			apiError = validateChildrenFieldsAndApplyDefaults(childrenPermissionMap, input.Children, formData, lockedStore)
			if apiError != service.NoError {
				return apiError.Error // rollback
			}

			parentChildSpec := constructItemsItemsForChildren(input.Children, itemID)
			insertItemItems(lockedStore, parentChildSpec)
			return lockedStore.ItemItems().After()
		})
	} else if formData.IsSet("no_score") || formData.IsSet("validation_type") {
		// results data of the task will be zeroed
		service.MustNotBeError(store.Results().Where("item_id = ?", itemID).
			UpdateColumn("result_propagation_state", "to_be_recomputed").Error())
		service.MustNotBeError(store.Results().Propagate())
	}
	return apiError, err
}

// constructUpdateItemChildTypeNonSkillValidator constructs a validator for the Children field that checks
// if a child's type is not 'Skill' when the items's type is not 'Skill'.
func constructUpdateItemChildTypeNonSkillValidator(itemType string,
	childrenInfoMap *map[int64]permissionAndType) validator.Func { // nolint:gocritic
	return validator.Func(func(fl validator.FieldLevel) bool {
		child := fl.Field().Interface().(itemChild)
		if itemType == skill {
			return true
		}
		return (*childrenInfoMap)[child.ItemID].Type != skill
	})
}
