package items

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/validator"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemChild struct {
	// required: true
	ItemID int64 `json:"item_id,string" sql:"column:child_item_id" validate:"set,child_item_id"`
	// default: 0
	Order int32 `json:"order" sql:"column:child_order"`
}

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

	// Nullable
	//
	// An optional comma-separated list of items' IDs to unlock (the current user should have `can_grant_view` >= 'content' on each)
	UnlockedItemIDs *string `json:"unlocked_item_ids" validate:"unlocked_item_ids"`
	// Nullable
	ScoreMinUnlock *int32 `json:"score_min_unlock"`
	// enum: All,Half,One,None
	// default: None
	ContestEnteringCondition string `json:"contest_entering_condition" validate:"oneof=All Half One None"`
	ContestMaxTeamSize       int32  `json:"contest_max_team_size"`
	TitleBarVisible          bool   `json:"title_bar_visible"`
	HasAttempts              bool   `json:"has_attempts"`
	// Nullable
	//
	// MySQL time (max value is 838:59:59)
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	Duration      *string `json:"duration" validate:"duration"`
	ShowUserInfos bool    `json:"show_user_infos"`
	UsesAPI       bool    `json:"uses_api"`
	// Nullable
	GroupCodeEnter *bool `json:"group_code_enter"`
}

type itemWithRequiredType struct {
	item `json:"item,squash"`
	// required: true
	// enum: Root,Category,Chapter,Task,Course
	Type string `json:"type" validate:"set,oneof=Root Category Chapter Task Course"`
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
	// Nullable fields are of pointer types
	itemWithRequiredType `json:"item,squash"`
	// `default_language_id` of the item
	// required: true
	LanguageID    int64 `json:"language_id" validate:"set,language_id"`
	newItemString `json:"string,squash"`

	// required: true
	ParentItemID int64 `json:"parent_item_id,string" validate:"set,parent_item_id"`
	// default: 0
	Order int32 `json:"order"`

	Children []itemChild `json:"children" validate:"children"`
}

func (in *NewItemRequest) canCreateItemsRelationsWithoutCycles(store *database.DataStore) bool {
	if len(in.Children) == 0 {
		return true
	}
	ids := make([]int64, len(in.Children)+1)
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
//   Creates an item with parameters from the input data with `items.default_language_id` = `language_id`.
//   Also it
//
//     * inserts a row into `items_strings` with given `language_id`, `title`, `image_url`, `subtitle`, `description`,
//
//     * gives full access to the item for the current user (creates a new `permissions_granted` row with: `item_id` = `items.id`,
//       `group_id` = `group_id` of the current user, `giver_group_id` = `users.group_id` of the current user,
//       `is_owner` = 1).
//
//     * adds new relations for the parent and (optionally) children items into `items_items` and propagates `permissions_generated`.
//
//   The user should have
//
//     * `can_edit` >= 'children' on the `parent_item_id`,
//     * `can_view` != 'none' on the `children` items (if any),
//     * `can_grant_view` >= 'content' on the `unlocked_item_ids` items (if any),
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
		registerAddItemValidators(formData, store, user)

		err = formData.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		err = store.WithNamedLock("items_items", 3*time.Second, func(lockedStore *database.DataStore) error {
			if !input.canCreateItemsRelationsWithoutCycles(lockedStore) {
				apiError = service.ErrForbidden(errors.New("an item cannot become an ancestor of itself"))
				return apiError.Error // rollback
			}

			// insertion
			itemID, err = srv.insertItem(lockedStore, user, formData, &input)
			return err
		})
		return err
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
func constructParentItemIDValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		hasAccess, err := store.Permissions().MatchingUserAncestors(user).WithWriteLock().
			Where("item_id = ?", fl.Field().Interface().(int64)).
			WherePermissionIsAtLeast("edit", "children").
			HasRows()
		service.MustNotBeError(err)
		return hasAccess
	})
}

// constructLanguageIDValidator constructs a validator for the LanguageID field.
// The validator checks that the language exists.
func constructLanguageIDValidator(store *database.DataStore) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		found, err := store.Languages().ByID(fl.Field().Interface().(int64)).WithWriteLock().HasRows()
		service.MustNotBeError(err)
		return found
	})
}

// constructUnlockedItemIDsValidator constructs a validator for the UnlockedItemIDs field.
// The validator checks that the user can grant 'content' view right on all the listed items (can_grant_view >= content).
func constructUnlockedItemIDsValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.Ptr { // nil
			return true
		}
		ids := strings.Split(field.String(), ",")
		int64IDs := make([]int64, 0, len(ids))
		for _, id := range ids {
			int64ID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return false
			}
			int64IDs = append(int64IDs, int64ID)
		}
		hasAccess, err := store.Items().CanGrantViewContentOnAll(user, int64IDs...)
		service.MustNotBeError(err)
		return hasAccess
	})
}

// constructChildrenValidator constructs a validator for the Children field.
// The validator checks that there are no duplicates in the list and
// all the children items are visible to the user (can_view != 'none').
func constructChildrenValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		children := fl.Field().Interface().([]itemChild)

		idsMap := make(map[int64]bool, len(children))
		ids := make([]int64, len(children))
		for index := range children {
			idsMap[children[index].ItemID] = true
			ids[index] = children[index].ItemID
		}
		if len(idsMap) != len(children) {
			return false
		}
		hasAccess, err := store.Items().AreAllVisible(user, ids...)
		service.MustNotBeError(err)
		return hasAccess
	})
}

func registerAddItemValidators(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("parent_item_id", constructParentItemIDValidator(store, user))
	formData.RegisterTranslation("parent_item_id",
		"should exist and the user should be able to manage its children")

	registerLanguageIDValidator(formData, store)

	registerChildrenValidator(formData, store, user)
	registerItemValidators(formData, store, user)
}

func registerLanguageIDValidator(formData *formdata.FormData, store *database.DataStore) {
	formData.RegisterValidation("language_id", constructLanguageIDValidator(store))
	formData.RegisterTranslation("language_id", "no such language")
}

func registerItemValidators(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("unlocked_item_ids", constructUnlockedItemIDsValidator(store, user))
	formData.RegisterTranslation("unlocked_item_ids",
		"all the IDs should exist and the user should have `can_grant_view` >= 'content' permission on each")
}

func registerChildrenValidator(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("children", constructChildrenValidator(store, user))
	formData.RegisterTranslation("children",
		"children IDs should be unique and each should be visible to the user")
}

func (srv *Service) insertItem(store *database.DataStore, user *database.User, formData *formdata.FormData,
	newItemRequest *NewItemRequest) (itemID int64, err error) {
	itemMap := formData.ConstructPartialMapForDB("itemWithRequiredType")
	stringMap := formData.ConstructPartialMapForDB("newItemString")

	err = store.RetryOnDuplicatePrimaryKeyError(func(s *database.DataStore) error {
		itemID = s.NewID()

		itemMap["id"] = itemID
		itemMap["default_language_id"] = newItemRequest.LanguageID
		service.MustNotBeError(s.Items().InsertMap(itemMap))

		service.MustNotBeError(s.PermissionsGranted().InsertMap(
			map[string]interface{}{
				"item_id":        itemID,
				"group_id":       user.GroupID,
				"giver_group_id": user.GroupID,
				"is_owner":       true,
			}))

		stringMap["id"] = s.NewID()
		stringMap["item_id"] = itemID
		stringMap["language_id"] = newItemRequest.LanguageID
		service.MustNotBeError(s.ItemStrings().InsertMap(stringMap))

		parentChildSpec := make([]*insertItemItemsSpec, 0, 1+len(newItemRequest.Children))
		parentChildSpec = append(parentChildSpec,
			&insertItemItemsSpec{ParentItemID: newItemRequest.ParentItemID, ChildItemID: itemID, Order: newItemRequest.Order})
		for _, child := range newItemRequest.Children {
			parentChildSpec = append(parentChildSpec,
				&insertItemItemsSpec{ParentItemID: itemID, ChildItemID: child.ItemID, Order: child.Order})
		}
		insertItemItems(s, parentChildSpec)
		return store.ItemItems().After()
	})
	return itemID, err
}

type insertItemItemsSpec struct {
	ParentItemID int64
	ChildItemID  int64
	Order        int32
}

func insertItemItems(store *database.DataStore, spec []*insertItemItemsSpec) {
	if len(spec) == 0 {
		return
	}

	var values = make([]interface{}, 0, len(spec)*4)

	for index := range spec {
		values = append(values, store.NewID(), spec[index].ParentItemID, spec[index].ChildItemID, spec[index].Order)
	}

	valuesMarks := strings.Repeat("(?, ?, ?, ?), ", len(spec)-1) + "(?, ?, ?, ?)"
	// nolint:gosec
	query := fmt.Sprintf("INSERT INTO `items_items` (`id`, `parent_item_id`, `child_item_id`, `child_order`) VALUES %s",
		valuesMarks)
	service.MustNotBeError(store.Exec(query, values...).Error())
}
