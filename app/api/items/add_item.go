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

type item struct {
	// Nullable
	URL *string `json:"url"`
	// default: false
	TeamsEditable bool `json:"teams_editable"`
	// default: false
	NoScore bool `json:"no_score"`
	// Nullable
	TextID                 *string `json:"text_id"`
	DisplayDetailsInParent bool    `json:"display_details_in_parent"`
	ReadOnly               bool    `json:"read_only"`
	// enum: forceYes,forceNo,default,
	FullScreen   string `json:"full_screen" validate:"omitempty,oneof=forceYes forceNo default"`
	HintsAllowed bool   `json:"hints_allowed"`
	FixedRanks   bool   `json:"fixed_ranks"`

	// enum: None,All,AllButOne,Categories,One,Manual
	// default: All
	ValidationType string `json:"validation_type" validate:"oneof=None All AllButOne Categories One Manual"`

	// enum: All,Half,One,None
	// default: None
	ContestEnteringCondition string    `json:"contest_entering_condition" validate:"oneof=All Half One None"`
	EnteringTimeMin          time.Time `json:"entering_time_min"`
	EnteringTimeMax          time.Time `json:"entering_time_max"`
	ContestMaxTeamSize       int32     `json:"contest_max_team_size"`
	TitleBarVisible          bool      `json:"title_bar_visible"`
	AllowsMultipleAttempts   bool      `json:"allows_multiple_attempts"`
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type" validate:"oneof=User Team"`
	// Nullable
	//
	// MySQL time (max value is 838:59:59)
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	Duration                *string `json:"duration" validate:"omitempty,duration"`
	ShowUserInfos           bool    `json:"show_user_infos"`
	UsesAPI                 bool    `json:"uses_api"`
	PromptToJoinGroupByCode bool    `json:"prompt_to_join_group_by_code"`
}

type itemWithRequiredType struct {
	item `json:"item,squash"`
	// Can be equal to 'Skill' only if the parent's type is 'Skill'
	// required: true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type" validate:"set,oneof=Chapter Task Course Skill,type_skill"`
}

// swagger:ignore
type newItemString struct {
	// required: true
	Title string `json:"title" validate:"set"`
	// Nullable
	ImageURL *string `json:"image_url"`
	// Nullable
	Subtitle *string `json:"subtitle"`
	// Nullable
	Description *string `json:"description"`
}

// NewItemRequest is the expected input for new created item
// swagger:model itemCreateRequest
type NewItemRequest struct {
	// `default_language_tag` of the item
	// required: true
	LanguageTag   string `json:"language_tag" validate:"set,language_tag"`
	newItemString `json:"string,squash"`

	// required: true
	ParentItemID int64 `json:"parent_item_id,string" validate:"set,parent_item_id,parent_item_type"`

	// Nullable fields are of pointer types
	itemWithRequiredType `json:"item,squash"`

	// enum: Undefined,Discovery,Application,Validation,Challenge
	// default: Undefined
	Category string `json:"category" validate:"oneof=Undefined Discovery Application Validation Challenge"`
	// default: 1
	ScoreWeight int8 `json:"score_weight"`
	// default: as_info
	// enum: none,as_info,as_content
	ContentViewPropagation string `json:"content_view_propagation" validate:"oneof=none as_info as_content"`
	// default: as_is
	// enum: use_content_view_propagation,as_content_with_descendants,as_is
	UpperViewLevelsPropagation string `json:"upper_view_levels_propagation" validate:"oneof=use_content_view_propagation as_content_with_descendants as_is"` // nolint:lll
	// default: true
	GrantViewPropagation bool `json:"grant_view_propagation"`
	// default: true
	WatchPropagation bool `json:"watch_propagation"`
	// default: true
	EditPropagation bool        `json:"edit_propagation"`
	Children        []itemChild `json:"children" validate:"children,children_allowed,dive,child_type_non_skill"`
}

func (in *NewItemRequest) canCreateItemsRelationsWithoutCycles(store *database.DataStore) bool {
	if len(in.Children) == 0 {
		return true
	}
	ids := make([]int64, len(in.Children))
	for index := range in.Children {
		if in.Children[index].ItemID == in.ParentItemID {
			return false
		}
		ids[index] = in.Children[index].ItemID
	}
	var count int64
	service.MustNotBeError(store.ItemAncestors().WithWriteLock().
		Where("child_item_id = ?", in.ParentItemID).
		Where("ancestor_item_id IN (?)", ids).Count(&count).Error())
	return count == 0
}

// swagger:operation POST /items items itemCreate
// ---
// summary: Create an item
// description: >
//
//   Creates an item with parameters from the input data with `items.default_language_tag` = `language_tag`.
//   Also it
//
//     * inserts a row into `items_strings` with given `language_tag`, `title`, `image_url`, `subtitle`, `description`,
//
//     * gives full access to the item for the current user (creates a new `permissions_granted` row with: `item_id` = `items.id`,
//       `group_id` = `group_id` of the current user, `source_group_id` = `users.group_id` of the current user, `origin` = 'self',
//       `is_owner` = 1).
//
//     * adds new relations for the parent and (optionally) children items into `items_items` and propagates `permissions_generated`.
//       (The only allowed parent-child relations are skills-*, chapter-task, chapter-course, chapter-chapter.
//       Otherwise the "bad request" error is returned.)
//
//     * (if `duration` is set) creates a participants group, links `contest_participants_group_id` to it,
//       and gives this group 'can_view:content' permission on the new item.
//
//   The user should have
//
//     * `can_edit` >= 'children' on the `parent_item_id`,
//     * `can_view` != 'none' on the `children` items (if any),
//
//   otherwise the "bad request" response is returned.
// parameters:
// - in: body
//   name: data
//   required: true
//   description: The item to create
//   schema:
//     "$ref": "#/definitions/itemCreateRequest"
// responses:
//   "201":
//     "$ref": "#/responses/createdWithIDResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) addItem(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	input := NewItemRequest{}
	formData := formdata.NewFormData(&input)

	apiError := service.NoError
	var itemID int64
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		var parentInfo parentItemInfo
		var childrenInfoMap map[int64]permissionAndType
		registerAddItemValidators(formData, store, user, &parentInfo, &childrenInfoMap)

		err = formData.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		return store.WithNamedLock("items_items", 3*time.Second, func(lockedStore *database.DataStore) error {
			if !input.canCreateItemsRelationsWithoutCycles(lockedStore) {
				apiError = service.ErrForbidden(errors.New("an item cannot become an ancestor of itself"))
				return apiError.Error // rollback
			}

			apiError = validateChildrenFieldsAndApplyDefaults(childrenInfoMap, input.Children, formData, lockedStore)
			if apiError != service.NoError {
				return apiError.Error // rollback
			}

			// insertion
			itemID = srv.insertItem(lockedStore, user, formData, &input)
			return nil
		})
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	response := struct {
		ItemID int64 `json:"id,string"`
	}{ItemID: itemID}
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(&response)))
	return service.NoError
}

// constructParentItemIDValidator constructs a validator for the ParentItemID field.
// The validator checks that the user has rights to manage the parent item's children (can_edit >= children).
func constructParentItemIDValidator(store *database.DataStore, user *database.User, parentInfo *parentItemInfo) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		err := store.Items().WhereUserHasPermissionOnItems(user, "edit", "children").
			Where("items.id = ?", fl.Field().Interface().(int64)).Select("items.type").
			Limit(1).Scan(&parentInfo).Error()
		if gorm.IsRecordNotFoundError(err) {
			return false
		}
		service.MustNotBeError(err)
		return true
	})
}

// constructParentItemTypeValidator constructs a validator checking that the parent item is not a Task or a Course.
func constructParentItemTypeValidator(parentInfo *parentItemInfo) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		return parentInfo.Type != course && parentInfo.Type != task
	})
}

// constructLanguageTagValidator constructs a validator for the LanguageTag field.
// The validator checks that the language exists.
func constructLanguageTagValidator(store *database.DataStore) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		found, err := store.Languages().ByTag(fl.Field().Interface().(string)).WithWriteLock().HasRows()
		service.MustNotBeError(err)
		return found
	})
}

// constructTypeSkillValidator constructs a validator for the Type field.
// The validator checks that the parent item's type is 'Skill' when the item's type is 'Skill'.
func constructTypeSkillValidator(parentInfo *parentItemInfo) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		if parentInfo.Type == "" || fl.Field().String() != skill {
			return true
		}
		return parentInfo.Type == skill
	})
}

// constructChildrenValidator constructs a validator for the Children field.
// The validator checks that there are no duplicates in the list and
// all the children items are visible to the user (can_view != 'none').
func constructChildrenValidator(store *database.DataStore, user *database.User,
	childrenInfoMap *map[int64]permissionAndType) validator.Func { // nolint:gocritic
	return validator.Func(func(fl validator.FieldLevel) bool {
		children := fl.Field().Interface().([]itemChild)

		if len(children) == 0 {
			return true
		}

		idsMap := make(map[int64]bool, len(children))
		ids := make([]int64, len(children))
		for index := range children {
			idsMap[children[index].ItemID] = true
			ids[index] = children[index].ItemID
		}
		if len(idsMap) != len(children) {
			return false
		}

		var childrenInfo []permissionAndType
		service.MustNotBeError(store.Items().
			Where("item_id IN (?)", ids).
			WhereUserHasViewPermissionOnItems(user, "info").
			WithWriteLock().
			Select("permissions.*, items.type").
			Scan(&childrenInfo).Error())

		*childrenInfoMap = make(map[int64]permissionAndType, len(childrenInfo))
		for index := range childrenInfo {
			(*childrenInfoMap)[childrenInfo[index].ItemID] = childrenInfo[index]
		}

		return len(*childrenInfoMap) == len(ids)
	})
}

// constructChildrenAllowedValidator constructs a validator checking that the new item can have children (is not a Task or a Course).
func constructChildrenAllowedValidator(
	defaultItemType string, childrenInfoMap *map[int64]permissionAndType) validator.Func { // nolint:gocritic
	return validator.Func(func(fl validator.FieldLevel) bool {
		if len(*childrenInfoMap) == 0 {
			return true
		}
		var itemType string
		if fl.Top().Elem().FieldByName("Type").IsValid() {
			itemType = fl.Top().Elem().FieldByName("Type").String()
		} else {
			itemType = defaultItemType
		}
		return itemType != task && itemType != course
	})
}

// constructChildTypeNonSkillValidator constructs a validator for the Children field that check
// if a child's type is not 'Skill' when the items's type is not 'Skill'.
func constructChildTypeNonSkillValidator(childrenInfoMap *map[int64]permissionAndType) validator.Func { // nolint:gocritic
	return validator.Func(func(fl validator.FieldLevel) bool {
		child := fl.Field().Interface().(itemChild)

		itemType := fl.Top().Elem().FieldByName("Type").String()
		if itemType == skill {
			return true
		}
		return (*childrenInfoMap)[child.ItemID].Type != skill
	})
}

type parentItemInfo struct {
	Type string
}

func registerAddItemValidators(formData *formdata.FormData, store *database.DataStore, user *database.User,
	parentInfo *parentItemInfo, childrenInfoMap *map[int64]permissionAndType) { // nolint:gocritic
	formData.RegisterValidation("parent_item_id", constructParentItemIDValidator(store, user, parentInfo))
	formData.RegisterTranslation("parent_item_id",
		"should exist and the user should be able to manage its children")
	formData.RegisterValidation("parent_item_type", constructParentItemTypeValidator(parentInfo))
	formData.RegisterTranslation("parent_item_type",
		"parent item cannot be Task or Course")

	registerLanguageTagValidator(formData, store)
	formData.RegisterValidation("type_skill", constructTypeSkillValidator(parentInfo))
	formData.RegisterTranslation("type_skill", "type can be equal to 'Skill' only if the parent item is a skill")
	registerChildrenValidator(formData, store, user, "", childrenInfoMap)
	formData.RegisterValidation("child_type_non_skill", constructChildTypeNonSkillValidator(childrenInfoMap))
	formData.RegisterTranslation("child_type_non_skill", "a skill cannot be a child of a non-skill item")
}

func registerLanguageTagValidator(formData *formdata.FormData, store *database.DataStore) {
	formData.RegisterValidation("language_tag", constructLanguageTagValidator(store))
	formData.RegisterTranslation("language_tag", "no such language")
}

func registerChildrenValidator(formData *formdata.FormData, store *database.DataStore, user *database.User,
	itemType string, childrenInfoMap *map[int64]permissionAndType) { // nolint:gocritic
	formData.RegisterValidation("children", constructChildrenValidator(store, user, childrenInfoMap))
	formData.RegisterTranslation("children",
		"children IDs should be unique and each should be visible to the user")

	formData.RegisterValidation("children_allowed", constructChildrenAllowedValidator(itemType, childrenInfoMap))
	formData.RegisterTranslation("children_allowed", "a task or a course cannot have children items")
}

func (srv *Service) insertItem(store *database.DataStore, user *database.User, formData *formdata.FormData,
	newItemRequest *NewItemRequest) (itemID int64) {
	itemMap := formData.ConstructPartialMapForDB("itemWithRequiredType")
	stringMap := formData.ConstructPartialMapForDB("newItemString")

	service.MustNotBeError(store.WithForeignKeyChecksDisabled(func(fkStore *database.DataStore) error {
		return fkStore.RetryOnDuplicatePrimaryKeyError(func(s *database.DataStore) error {
			itemID = s.NewID()

			itemMap["id"] = itemID
			itemMap["default_language_tag"] = newItemRequest.LanguageTag
			return s.Items().InsertMap(itemMap)
		})
	}))

	if itemMap["duration"] != nil {
		participantsGroupID := createContestParticipantsGroup(store, itemID)
		service.MustNotBeError(store.Items().ByID(itemID).
			UpdateColumn("contest_participants_group_id", participantsGroupID).Error())
	}

	service.MustNotBeError(store.PermissionsGranted().InsertMap(
		map[string]interface{}{
			"item_id":         itemID,
			"group_id":        user.GroupID,
			"source_group_id": user.GroupID,
			"origin":          "self",
			"is_owner":        true,
		}))

	stringMap["item_id"] = itemID
	stringMap["language_tag"] = newItemRequest.LanguageTag
	service.MustNotBeError(store.ItemStrings().InsertMap(stringMap))

	if !formData.IsSet("category") {
		newItemRequest.Category = undefined
	}
	if !formData.IsSet("score_weight") {
		newItemRequest.ScoreWeight = 1
	}

	for index := range newItemRequest.Children {
		if !formData.IsSet(fmt.Sprintf("children[%d].category", index)) {
			newItemRequest.Children[index].Category = undefined
		}
		if !formData.IsSet(fmt.Sprintf("children[%d].score_weight", index)) {
			newItemRequest.Children[index].ScoreWeight = 1
		}
	}

	var order int32
	service.MustNotBeError(store.ItemItems().WithWriteLock().
		Where("parent_item_id = ?", newItemRequest.ParentItemID).
		PluckFirst("IFNULL(MAX(`child_order`), 0)+1", &order).Error())

	parentChildSpec := make([]*insertItemItemsSpec, 0, 1+len(newItemRequest.Children))
	parentChildSpec = append(parentChildSpec,
		&insertItemItemsSpec{
			ParentItemID: newItemRequest.ParentItemID, ChildItemID: itemID,
			Order:    order,
			Category: newItemRequest.Category, ScoreWeight: newItemRequest.ScoreWeight,
			ContentViewPropagation: asInfo, UpperViewLevelsPropagation: asIs,
			GrantViewPropagation: true, WatchPropagation: true, EditPropagation: true,
		})
	parentChildSpec = append(parentChildSpec,
		constructItemsItemsForChildren(newItemRequest.Children, itemID)...)
	insertItemItems(store, parentChildSpec)
	service.MustNotBeError(store.ItemItems().After())

	return itemID
}
