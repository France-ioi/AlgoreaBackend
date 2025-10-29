package items

import (
	"errors"
	"net/http"

	"github.com/France-ioi/validator"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// ItemWithDefaultLanguageTag represents common item fields plus 'default_language_tag'.
type ItemWithDefaultLanguageTag struct {
	Item `json:"item,squash"`

	// new `default_language_tag` of the item can only be set to a language
	// for that an `items_strings` row exists
	// minLength: 1
	// maxLength: 6
	DefaultLanguageTag string `json:"default_language_tag" validate:"min=1,max=6"`
}

// updateItemRequest is the expected input for item updating
// swagger:model itemEditRequest
type updateItemRequest struct {
	ItemWithDefaultLanguageTag `json:"item,squash"`

	Children []itemChild `json:"children" validate:"children,children_allowed,dive,child_type_non_skill"`

	childrenIDsCache []int64
}

func (in *updateItemRequest) childrenIDs() []int64 {
	if in.childrenIDsCache != nil {
		return in.childrenIDsCache
	}
	in.childrenIDsCache = make([]int64, len(in.Children))
	for index := range in.Children {
		in.childrenIDsCache[index] = in.Children[index].ItemID
	}
	return in.childrenIDsCache
}

func (in *updateItemRequest) checkItemsRelationsCycles(store *database.DataStore, itemID int64) bool {
	if len(in.Children) == 0 {
		return true
	}
	ids := in.childrenIDs()
	for _, id := range ids {
		if id == itemID {
			return false
		}
	}
	var count int64
	service.MustNotBeError(store.ItemAncestors().WithExclusiveWriteLock().
		Where("child_item_id = ?", itemID).
		Where("ancestor_item_id IN (?)", ids).Count(&count).Error())
	return count == 0
}

// swagger:operation PUT /items/{item_id} items itemUpdate
//
//	---
//	summary: Update an item
//	description: >
//
//		Changes the given item's properties using the input data.
//
//
//		If the `children` array is given, the service removes relations between the item and its former children,
//		inserts relations between the item and its new children into `items_items`, and propagates `permissions_generated`.
//		(The only allowed parent-child relations are skills-*, chapter-task, chapter-chapter.
//		Otherwise, the "bad request" error is returned.)
//
//
//		If `requires_explicit_entry` is being set to true and `participants_group_id` is NULL,
//		the service creates a participants group, links `participants_group_id` to it,
//		and gives this group 'can_view:content' permission on the new item.
//
//
//		The user should have
//
//			* `can_view` >= 'content' on the item, otherwise the "forbidden" response is returned;
//			* `can_edit` >= 'children' on the item to edit children or `can_edit` >= 'all' to edit the item's properties,
//				otherwise the "forbidden" response is returned;
//			* `can_view` != 'none' on the `children` items (if any), otherwise the "bad request"
//				response is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- in: body
//			name: data
//			required: true
//			description: New item property values
//			schema:
//				"$ref": "#/definitions/itemEditRequest"
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
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
func (srv *Service) updateItem(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	rawRequestData, err := service.ResolveJSONBodyIntoMap(httpRequest)
	service.MustNotBeError(err)

	var propagationsToRun []string

	err = store.InTransaction(func(store *database.DataStore) error {
		input := updateItemRequest{}
		formData := formdata.NewFormData(&input)
		var itemInfo struct {
			ParticipantsGroupID   *int64
			Type                  string
			CanEditGeneratedValue int
			Duration              *string
			RequiresExplicitEntry bool
		}
		err = store.Permissions().MatchingUserAncestors(user).WithExclusiveWriteLock().
			Joins("JOIN items ON items.id = item_id").
			Where("item_id = ?", itemID).
			HavingMaxPermissionAtLeast("view", "content").
			HavingMaxPermissionAtLeast("edit", "children").
			Select(`
				items.participants_group_id, items.type, MAX(can_edit_generated_value) AS can_edit_generated_value,
				items.duration, items.requires_explicit_entry`).
			Group("item_id").
			Scan(&itemInfo).Error()

		if gorm.IsRecordNotFoundError(err) {
			return service.ErrForbidden(errors.New("no access rights to edit the item")) // rollback
		}
		service.MustNotBeError(err)

		var childrenInfoMap map[int64]permissionAndType
		var oldPropagationLevelsMap map[int64]*itemsRelationData

		registerChildrenValidator(formData, store, user, itemInfo.Type, &childrenInfoMap, &oldPropagationLevelsMap, &itemID)
		formData.RegisterValidation("child_type_non_skill", constructUpdateItemChildTypeNonSkillValidator(itemInfo.Type, &childrenInfoMap))
		formData.RegisterTranslation("child_type_non_skill", "a skill cannot be a child of a non-skill item")
		formData.RegisterValidation("cannot_be_set_for_skills", constructUpdateItemCannotBeSetForSkillsValidator(itemInfo.Type))
		formData.RegisterTranslation("cannot_be_set_for_skills", "cannot be set for skill items")
		formData.RegisterValidation("duration_requires_explicit_entry",
			constructUpdateItemDurationRequiresExplicitEntryValidator(formData, itemInfo.Duration, itemInfo.RequiresExplicitEntry))
		formData.RegisterTranslation("duration_requires_explicit_entry", "requires_explicit_entry should be true when the duration is not null")
		formData.RegisterValidation("options", constructItemOptionsValidator())
		formData.RegisterTranslation("null|options", "options should be a valid JSON or null")

		err = formData.ParseMapData(rawRequestData)
		if err != nil {
			return service.ErrInvalidRequest(err) // rollback
		}

		itemData := formData.ConstructPartialMapForDB("ItemWithDefaultLanguageTag")
		if len(itemData) == 0 && !formData.IsSet("children") {
			return nil // Nothing to do
		}

		if len(itemData) > 0 &&
			itemInfo.CanEditGeneratedValue < store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all") {
			return service.ErrForbidden(errors.New("no access rights to edit the item's properties")) // rollback
		}

		err = updateItemInDB(itemData, itemInfo.ParticipantsGroupID, store, itemID)
		if err != nil {
			return err // rollback
		}

		propagationsToRun, err = updateChildrenAndRunListeners(
			formData,
			store,
			itemID,
			&input,
			childrenInfoMap,
			oldPropagationLevelsMap,
		)
		return err
	})

	service.MustNotBeError(err)

	service.SchedulePropagation(store, srv.GetPropagationEndpoint(), propagationsToRun)

	// response
	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess[*struct{}](nil)))
	return nil
}

func updateItemInDB(
	itemData map[string]interface{}, participantsGroupID *int64, store *database.DataStore, itemID int64,
) error {
	if itemData["requires_explicit_entry"] == true && participantsGroupID == nil {
		createdParticipantsGroupID := createParticipantsGroupForItemRequiringExplicitEntry(store, itemID)
		itemData["participants_group_id"] = createdParticipantsGroupID
	}

	err := store.Items().Where("id = ?", itemID).UpdateColumn(itemData).Error()
	if err != nil {
		if database.IsDuplicateEntryError(err) {
			return service.ErrForbidden(formdata.FieldErrorsError{"text_id": []string{
				"text_id must be unique",
			}})
		} else if database.IsKindOfNoReferencedRowError(err) {
			return service.ErrInvalidRequest(formdata.FieldErrorsError{"default_language_tag": []string{
				"default language should exist and there should be item's strings in this language",
			}})
		}
	}
	service.MustNotBeError(err)

	return nil
}

func updateChildrenAndRunListeners(
	formData *formdata.FormData,
	store *database.DataStore,
	itemID int64,
	input *updateItemRequest,
	childrenPermissionMap map[int64]permissionAndType,
	oldPropagationLevelsMap map[int64]*itemsRelationData,
) (propagationsToRun []string, err error) {
	if formData.IsSet("children") {
		err = store.ItemItems().WithItemsRelationsLock(func(lockedStore *database.DataStore) error {
			deleteStatement := lockedStore.ItemItems().Where("parent_item_id = ?", itemID)
			newChildrenIDs := input.childrenIDs()
			if len(newChildrenIDs) > 0 {
				deleteStatement = deleteStatement.Where("child_item_id NOT IN(?)", newChildrenIDs)
			}
			service.MustNotBeError(deleteStatement.Delete().Error())

			if !input.checkItemsRelationsCycles(lockedStore, itemID) {
				return service.ErrForbidden(errors.New("an item cannot become an ancestor of itself")) // rollback
			}

			service.MustNotBeError(validateChildrenFieldsAndApplyDefaults(
				childrenPermissionMap, input.Children, formData, oldPropagationLevelsMap, lockedStore))

			parentChildSpec := constructItemsItemsForChildren(input.Children, itemID)
			insertItemItems(lockedStore, parentChildSpec)

			service.MustNotBeError(store.ItemItems().CreateNewAncestors())

			return nil
		})

		propagationsToRun = []string{"permissions", "results"}
	} else if formData.IsSet("no_score") || formData.IsSet("validation_type") {
		// results data of the task will be zeroed
		service.MustNotBeError(
			store.Exec("INSERT INTO results_propagate ?",
				store.Results().Where("item_id = ?", itemID).
					Select("participant_id, attempt_id, item_id, 'to_be_recomputed' AS state").QueryExpr()).Error())

		propagationsToRun = []string{"results"}
	}

	return propagationsToRun, err
}

// constructUpdateItemChildTypeNonSkillValidator constructs a validator for the Children field that checks
// if a child's type is not 'Skill' when the items's type is not 'Skill'.
func constructUpdateItemChildTypeNonSkillValidator(
	itemType string,
	childrenInfoMap *map[int64]permissionAndType, //nolint:gocritic // we need the pointer as the constructor is called before the map is set
) validator.Func {
	return func(fl validator.FieldLevel) bool {
		//nolint:forcetypeassert // the validator is registered only for elements of Children slice of type itemChild
		child := fl.Field().Interface().(itemChild)
		if itemType == skill {
			return true
		}
		return (*childrenInfoMap)[child.ItemID].Type != skill
	}
}

// constructUpdateItemCannotBeSetForSkillsValidator constructs a validator checking that the fields is not set for skill items.
func constructUpdateItemCannotBeSetForSkillsValidator(itemType string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return fl.Field().IsZero() || itemType != skill
	}
}

// constructUpdateItemDurationRequiresExplicitEntryValidator constructs a validator for the RequiresExplicitEntry field.
// The validator checks that when the duration is given and is not null or is not given while its previous value is not null,
// the field is true.
func constructUpdateItemDurationRequiresExplicitEntryValidator(
	formData *formdata.FormData, duration *string, requiresExplicitEntry bool,
) validator.Func {
	return func(fl validator.FieldLevel) bool {
		//nolint:forcetypeassert // the validator is registered only for fields inside Item
		data := fl.Parent().Addr().Interface().(*Item)
		var changed bool
		if formData.IsSet("duration") {
			if (duration == nil) != (data.Duration == nil) {
				changed = true
			}
			duration = data.Duration
		}
		if formData.IsSet("requires_explicit_entry") {
			if requiresExplicitEntry != data.RequiresExplicitEntry {
				changed = true
			}
			requiresExplicitEntry = data.RequiresExplicitEntry
		}
		return !changed || requiresExplicitEntry || duration == nil
	}
}
