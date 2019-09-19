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
	ItemID int64 `json:"item_id,string" sql:"column:item_child_id" validate:"set,child_item_id"`
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
	TextID *string `json:"text_id"`
	// Nullable
	CustomChapter          *bool `json:"custom_chapter"`
	DisplayDetailsInParent bool  `json:"display_details_in_parent"`
	ReadOnly               bool  `json:"read_only"`
	// enum: forceYes,forceNo,default,
	FullScreen     string `json:"full_screen" validate:"omitempty,oneof=forceYes forceNo default"`
	ShowDifficulty bool   `json:"show_difficulty"`
	ShowSource     bool   `json:"show_source"`
	HintsAllowed   bool   `json:"hints_allowed"`
	FixedRanks     bool   `json:"fixed_ranks"`

	// enum: None,All,AllButOne,Categories,One,Manual
	// default: All
	ValidationType string `json:"validation_type" validate:"oneof=None All AllButOne Categories One Manual"`

	// Nullable
	ValidationMin *int32 `json:"validation_min"`
	// Nullable
	//
	// An optional comma-separated list of items' IDs to unlock (each must be owned/managed by the current user)
	UnlockedItemIDs *string `json:"unlocked_item_ids" sql:"column:item_unlocked_id" validate:"unlocked_item_ids"`
	// Nullable
	ScoreMinUnlock *int32 `json:"score_min_unlock"`
	// Nullable
	// enum: All,Half,One,None
	TeamMode *string `json:"team_mode" validate:"oneof=All Half One None"`
	// Nullable
	//
	// Should be owned by the current user
	TeamInGroupID   *int64 `json:"team_in_group_id" validate:"team_in_group_id"`
	TeamMaxMembers  int32  `json:"team_max_members"`
	TitleBarVisible bool   `json:"title_bar_visible"`
	HasAttempts     bool   `json:"has_attempts"`
	// Nullable
	AccessOpenDate *time.Time `json:"access_open_date"`
	// Nullable
	//
	// MySQL time (max value is 838:59:59)
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	Duration *string `json:"duration" validate:"duration"`
	// Nullable
	EndContestDate *time.Time `json:"end_contest_date"`
	// Nullable
	// enum: Running,Analysis,Closed
	ContestPhase  *string `json:"contest_phase" validate:"oneof=Running Analysis Closed"`
	ShowUserInfos bool    `json:"show_user_infos"`
	// Nullable
	Level   *int32 `json:"level"`
	UsesAPI bool   `json:"uses_api"`
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

// groupItemData creates a map containing the db data to be inserted into the groups_items table
func (in *NewItemRequest) groupItemData(groupItemID, userID, groupID, itemID int64) map[string]interface{} {
	return map[string]interface{}{
		"id":               groupItemID,
		"item_id":          itemID,
		"group_id":         groupID,
		"user_created_id":  userID,
		"full_access_date": database.Now(),
		"owner_access":     true,
		"manager_access":   true,
		// as the owner gets full access, there is no need to request parents' access to get the actual access level
		"cached_full_access_date": database.Now(),
		"cached_full_access":      true,
	}
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
		Where("item_child_id = ?", in.ParentItemID).
		Where("item_ancestor_id IN (?)", ids).Count(&count).Error())
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
//     * gives full access to the item for the current user (creates a new `groups_items` row with: `item_id` = `items.id`,
//       `group_id` = `group_self_id` of the current user, `user_created_id` = `users.id` of the current user,
//       `full_access_date` = now(), `cached_full_access_date` = now(), `cached_full_access` = 1, `owner_access` = 1,
//       `manager_access` = 1).
//
//     * adds new relations for the parent and (optionally) children items into `items_items` and propagates `groups_items`.
//
//   The user should be an owner/manager of
//
//     * the `parent_item_id`,
//     * `children` items (if any),
//     * `unlocked_item_ids` items (if any),
//
//   and be an owner of `team_in_group_id` (if given),
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
// The validator checks that the user has access rights to manage the parent item (owner_access or manager_access).
func constructParentItemIDValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		hasAccess, err := store.Items().HasManagerAccess(user, fl.Field().Interface().(int64))
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

// constructTeamInGroupIDValidator constructs a validator for the TeamInGroupID field.
// The validator checks that the group in the TeamInGroupID field is owned by the user.
func constructTeamInGroupIDValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.Ptr { // nil
			return true
		}
		found, err := store.Groups().
			OwnedBy(user).Where("groups.id = ?", field.Int()).
			WithWriteLock().HasRows()
		service.MustNotBeError(err)
		return found
	})
}

// constructUnlockedItemIDsValidator constructs a validator for the UnlockedItemIDs field.
// The validator checks that the user has access rights to manage all the listed items (owner_access or manager_access).
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
		hasAccess, err := store.Items().HasManagerAccess(user, int64IDs...)
		service.MustNotBeError(err)
		return hasAccess
	})
}

// constructChildrenValidator constructs a validator for the Children field.
// The validator checks that there are no duplicates in the list and
// the user has access rights to manage all the listed items (owner_access or manager_access).
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
		hasAccess, err := store.Items().HasManagerAccess(user, ids...)
		service.MustNotBeError(err)
		return hasAccess
	})
}

func registerAddItemValidators(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("parent_item_id", constructParentItemIDValidator(store, user))
	formData.RegisterTranslation("parent_item_id",
		"should exist and the user should have manager/owner access to it")

	registerLanguageIDValidator(formData, store)

	registerChildrenValidator(formData, store, user)
	registerItemValidators(formData, store, user)
}

func registerLanguageIDValidator(formData *formdata.FormData, store *database.DataStore) {
	formData.RegisterValidation("language_id", constructLanguageIDValidator(store))
	formData.RegisterTranslation("language_id", "no such language")
}

func registerItemValidators(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("team_in_group_id", constructTeamInGroupIDValidator(store, user))
	formData.RegisterTranslation("team_in_group_id", "should exist and be owned by the user")

	formData.RegisterValidation("unlocked_item_ids", constructUnlockedItemIDsValidator(store, user))
	formData.RegisterTranslation("unlocked_item_ids",
		"all the IDs should exist and the user should have manager/owner access to them")
}

func registerChildrenValidator(formData *formdata.FormData, store *database.DataStore, user *database.User) {
	formData.RegisterValidation("children", constructChildrenValidator(store, user))
	formData.RegisterTranslation("children",
		"children IDs should be unique and the user should have manager/owner access to them")
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

		// user.SelfGroupID is not null since we have successfully passed the validation
		service.MustNotBeError(s.GroupItems().InsertMap(newItemRequest.groupItemData(s.NewID(), user.ID, *user.SelfGroupID, itemID)))

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
	query := fmt.Sprintf("INSERT INTO `items_items` (`id`, `item_parent_id`, `item_child_id`, `child_order`) VALUES %s",
		valuesMarks)
	service.MustNotBeError(store.Exec(query, values...).Error())
}
