package items

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID   types.OptionalString `json:"id"`
	Type types.RequiredString `json:"type"`

	Strings []struct {
		LanguageID  types.RequiredString `json:"language_id"`
		Title       types.RequiredString `json:"title"`
		ImageURL    types.OptNullString  `json:"image_url"`
		Subtitle    types.OptNullString  `json:"subtitle"`
		Description types.OptNullString  `json:"description"`
	} `json:"strings"`

	Parents []struct {
		ID    types.RequiredString `json:"id"`
		Order types.RequiredInt64  `json:"order"` // actually it should be int32
	} `json:"parents"`
}

// Bind validates the request body attributes
func (in *NewItemRequest) Bind(r *http.Request) error {
	if len(in.Strings) != 1 {
		return errors.New("exactly one string per item is supported at the moment")
	}
	if len(in.Parents) != 1 {
		return errors.New("exactly one parent item is supported at the moment")
	}
	if err := types.Validate([]string{"id", "type"}, &in.ID, &in.Type); err != nil {
		return err
	}
	if in.ID.Set {
		if _, err := strconv.ParseInt(in.ID.Value, 10, 64); err != nil {
			return errors.New("'id' should be a number")
		}
	}
	if _, err := strconv.ParseInt(in.Strings[0].LanguageID.Value, 10, 64); err != nil {
		return errors.New("'strings[0].language_id' should be a number")
	}
	if _, err := strconv.ParseInt(in.Parents[0].ID.Value, 10, 64); err != nil {
		return errors.New("'parents[0].id' should be a number")
	}
	return nil
}

func (in *NewItemRequest) itemData() *database.Item {
	id, err := strconv.ParseInt(in.ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	languageID, err := strconv.ParseInt(in.Strings[0].LanguageID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	return &database.Item{
		ID:                *types.NewInt64(id),
		Type:              in.Type.String,
		DefaultLanguageID: *types.NewInt64(languageID),
		TeamsEditable:     *types.NewBool(false), // has no db default at the moment, so must be set
		NoScore:           *types.NewBool(false), // has no db default at the moment, so must be set
	}
}

func (in *NewItemRequest) groupItemData(id int64, userID int64, groupID int64) *database.GroupItem {
	itemID, err := strconv.ParseInt(in.ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	return &database.GroupItem{
		ID:             *types.NewInt64(id),
		ItemID:         *types.NewInt64(itemID),
		GroupID:        *types.NewInt64(groupID),
		CreatorUserID:  *types.NewInt64(userID),
		FullAccessDate: *types.NewDatetime(time.Now()),
		OwnerAccess:    *types.NewBool(true),
		ManagerAccess:  *types.NewBool(true),
		// as the owner gets full access, there is no need to request parents' access to get the actual access level
		CachedFullAccessDate: *types.NewDatetime(time.Now()),
		CachedFullAccess:     *types.NewBool(true),
	}
}

func (in *NewItemRequest) stringData(id int64) *database.ItemString {
	itemID, err := strconv.ParseInt(in.ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	languageID, err := strconv.ParseInt(in.Strings[0].LanguageID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	return &database.ItemString{
		ID:          *types.NewInt64(id),
		ItemID:      *types.NewInt64(itemID),
		LanguageID:  *types.NewInt64(languageID),
		Title:       in.Strings[0].Title.String,
		ImageURL:    in.Strings[0].ImageURL.String,
		Subtitle:    in.Strings[0].Subtitle.String,
		Description: in.Strings[0].Description.String,
	}
}
func (in *NewItemRequest) itemItemData(id int64) *database.ItemItem {
	itemID, err := strconv.ParseInt(in.ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	parentItemID, err := strconv.ParseInt(in.Parents[0].ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()
	return &database.ItemItem{
		ID:           *types.NewInt64(id),
		ChildItemID:  *types.NewInt64(itemID),
		Order:        in.Parents[0].Order.Int64,
		ParentItemID: *types.NewInt64(parentItemID),
	}
}

func (srv *Service) addItem(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	// validate input (could be moved to JSON validation later)
	input := &NewItemRequest{}
	if err = render.Bind(r, input); err != nil {
		return service.ErrInvalidRequest(err)
	}
	parentItemID, err := strconv.ParseInt(input.Parents[0].ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()

	// check permissions
	if ret := srv.checkPermission(user, parentItemID); ret != service.NoError {
		return ret
	}

	// insertion
	if err = srv.insertItem(user, input); err != nil {
		return service.ErrInvalidRequest(err)
	}

	id, err := strconv.ParseInt(input.ID.Value, 10, 64)
	service.MustNotBeError(err) // we have checked this in Bind()

	// response
	response := struct {
		ItemID string `json:"ID"`
	}{strconv.FormatInt(id, 10)}
	if err = render.Render(w, r, service.CreationSuccess(&response)); err != nil {
		return service.ErrUnexpected(err)
	}
	return service.NoError
}

func (srv *Service) insertItem(user *database.User, input *NewItemRequest) error {
	if !input.ID.Set {
		input.ID.Value = strconv.FormatInt(srv.Store.NewID(), 10)
		input.ID.Set = true
		input.ID.Null = false
	}

	return srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.Items().Insert(input.itemData()))
		userSelfGroupID, _ := user.SelfGroupID() // the user has been already loaded in checkPermission()
		service.MustNotBeError(store.GroupItems().Insert(input.groupItemData(store.NewID(), user.UserID, userSelfGroupID)))
		service.MustNotBeError(store.ItemStrings().Insert(input.stringData(store.NewID())))
		return store.ItemItems().Insert(input.itemItemData(store.NewID()))
	})
}

func (srv *Service) checkPermission(user *database.User, parentItemID int64) service.APIError {
	// can add a parent only if manager of that parent
	found, hasAccess, err := srv.Store.Items().HasManagerAccess(user, parentItemID)
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)
	if !found {
		return service.ErrForbidden(errors.New("cannot find the parent item"))
	}
	if !hasAccess {
		return service.ErrForbidden(errors.New("insufficient access on the parent item"))
	}
	return service.NoError
}
