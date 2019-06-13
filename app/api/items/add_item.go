package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemChild struct {
	ItemID int64 `json:"item_id,string" sql:"column:idItemChild" validate:"required,child_item_id"`
	Order  int32 `json:"order" sql:"column:iChildOrder"`
}

type item struct {
	URL                    *string `json:"url" sql:"column:sUrl"`
	TeamsEditable          bool    `json:"teams_editable" sql:"column:bTeamsEditable"`
	NoScore                bool    `json:"no_score" sql:"column:bNoScore"`
	TextID                 *string `json:"text_id" sql:"column:sTextId"`
	CustomChapter          *bool   `json:"custom_chapter" sql:"column:bCustomChapter"`
	DisplayDetailsInParent bool    `json:"display_details_in_parent" sql:"column:bDisplayDetailsInParent"`
	ReadOnly               bool    `json:"read_only" sql:"column:bReadOnly"`
	FullScreen             string  `json:"full_screen" sql:"column:sFullScreen" validate:"omitempty,oneof=forceYes forceNo default"`
	ShowDifficulty         bool    `json:"show_difficulty" sql:"column:bShowDifficulty"`
	ShowSource             bool    `json:"show_source" sql:"column:bShowSource"`
	HintsAllowed           bool    `json:"hints_allowed" sql:"column:bHintsAllowed"`
	FixedRanks             bool    `json:"fixed_ranks" sql:"column:bFixedRanks"`

	ValidationType string `json:"validation_type" sql:"column:sValidationType" validate:"oneof=None All AllButOne Categories One Manual"`

	ValidationMin   *int32     `json:"validation_min" sql:"column:iValidationMin"`
	UnlockedItemIDs *string    `json:"unlocked_item_ids" sql:"column:idItemUnlocked" validate:"unlocked_item_ids"`
	ScoreMinUnlock  *int32     `json:"score_min_unlock" sql:"column:iScoreMinUnlock"`
	TeamMode        *string    `json:"team_mode" sql:"column:sTeamMode" validate:"oneof=All Half One None"`
	TeamInGroupID   *int64     `json:"team_in_group_id" sql:"column:idTeamInGroup" validate:"team_in_group_id"`
	TeamMaxMembers  int32      `json:"team_max_members" sql:"column:iTeamMaxMembers"`
	TitleBarVisible bool       `json:"title_bar_visible" sql:"column:bTitleBarVisible"`
	HasAttempts     bool       `json:"has_attempts" sql:"column:bHasAttempts"`
	AccessOpenDate  *time.Time `json:"access_open_date" sql:"column:sAccessOpenDate"`
	Duration        *string    `json:"duration" sql:"column:sDuration" validate:"duration"`
	EndContestDate  *time.Time `json:"end_contest_date" sql:"column:sEndContestDate"`
	ContestPhase    *string    `json:"contest_phase" sql:"column:sContestPhase" validate:"oneof=Running Analysis Closed"`
	ShowUserInfos   bool       `json:"show_user_infos" sql:"column:bShowUserInfos"`
	Level           *int32     `json:"level" sql:"column:iLevel"`
	UsesAPI         bool       `json:"uses_api" sql:"column:bUsesAPI"`
	GroupCodeEnter  *bool      `json:"group_code_enter" sql:"column:groupCodeEnter"`
}

type itemWithRequiredType struct {
	Item item   `json:"item,squash"`
	Type string `json:"type" validate:"required,oneof=Root Category Chapter Task Course" sql:"column:sType"`
}

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	// Nullable fields are of pointer types
	Item       itemWithRequiredType `json:"item,squash"`
	LanguageID int64                `json:"language_id" validate:"required,language_id"`
	String     struct {
		Title       string  `json:"title" validate:"required" sql:"column:sTitle"`
		ImageURL    *string `json:"image_url" sql:"column:sImageUrl"`
		Subtitle    *string `json:"subtitle" sql:"column:sSubtitle"`
		Description *string `json:"description" sql:"column:sDescription"`
	} `json:"string,squash"`

	ParentItemID int64 `json:"parent_item_id,string" validate:"required,parent_item_id"`
	Order        int32 `json:"order"`

	Children []itemChild `json:"children" validate:"children"`
}

// groupItemData creates a map containing the db data to be inserted into the groups_items table
func (in *NewItemRequest) groupItemData(groupItemID, userID, groupID, itemID int64) map[string]interface{} {
	return map[string]interface{}{
		"ID":              groupItemID,
		"idItem":          itemID,
		"idGroup":         groupID,
		"idUserCreated":   userID,
		"sFullAccessDate": gorm.Expr("NOW()"),
		"bOwnerAccess":    true,
		"bManagerAccess":  true,
		// as the owner gets full access, there is no need to request parents' access to get the actual access level
		"sCachedFullAccessDate": gorm.Expr("NOW()"),
		"bCachedFullAccess":     true,
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
		Where("idItemChild = ?", in.ParentItemID).
		Where("idItemAncestor IN (?)", ids).Count(&count).Error())
	return count == 0
}

func (srv *Service) addItem(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)
	err = user.Load()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

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
		ItemID int64 `json:"ID,string"`
	}{ItemID: itemID}
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(&response)))
	return service.NoError
}

// constructParentItemIDValidator constructs a validator for the ParentItemID field.
// The validator checks that the user has access rights to manage the parent item (bOwnerAccess or bManagerAccess).
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
		found, err := store.Groups().
			OwnedBy(user).Where("groups.ID = ?", fl.Field().Interface().(int64)).
			WithWriteLock().HasRows()
		service.MustNotBeError(err)
		return found
	})
}

// constructUnlockedItemIDsValidator constructs a validator for the UnlockedItemIDs field.
// The validator checks that the user has access rights to manage all the listed items (bOwnerAccess or bManagerAccess).
func constructUnlockedItemIDsValidator(store *database.DataStore, user *database.User) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		ids := strings.Split(fl.Field().Interface().(string), ",")
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
// the user has access rights to manage all the listed items (bOwnerAccess or bManagerAccess).
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
	itemMap := formData.ConstructPartialMapForDB("Item")
	stringMap := formData.ConstructPartialMapForDB("String")

	err = store.RetryOnDuplicatePrimaryKeyError(func(s *database.DataStore) error {
		itemID = s.NewID()

		itemMap["ID"] = itemID
		itemMap["idDefaultLanguage"] = newItemRequest.LanguageID
		service.MustNotBeError(s.Items().InsertMap(itemMap))

		userSelfGroupID, _ := user.SelfGroupID() // the user has been already loaded in addItem()
		service.MustNotBeError(s.GroupItems().InsertMap(newItemRequest.groupItemData(s.NewID(), user.UserID, userSelfGroupID, itemID)))

		stringMap["ID"] = s.NewID()
		stringMap["idItem"] = itemID
		stringMap["idLanguage"] = newItemRequest.LanguageID
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
	query := fmt.Sprintf("INSERT INTO `items_items` (`ID`, `idItemParent`, `idItemChild`, `iChildOrder`) VALUES %s",
		valuesMarks)
	service.MustNotBeError(store.Exec(query, values...).Error())
}
