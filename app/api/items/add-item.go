package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID   types.OptionalInt64  `json:"id"`
	Type types.RequiredString `json:"type"`

	Strings []struct {
		LanguageID  types.RequiredInt64  `json:"language_id"`
		Title       types.RequiredString `json:"title"`
		ImageURL    types.OptNullString  `json:"image_url"`
		Subtitle    types.OptNullString  `json:"subtitle"`
		Description types.OptNullString  `json:"description"`
	} `json:"strings"`

	Parents []struct {
		ID    types.RequiredInt64 `json:"id"`
		Order types.RequiredInt32 `json:"order"`
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
	return types.Validate([]string{"id", "type", "strings[0].language_id", "parents[0].id"},
		&in.ID, &in.Type, &in.Strings[0].LanguageID, &in.Parents[0].ID)
}

func (in *NewItemRequest) itemData() *database.Item {
	return &database.Item{
		ID:                in.ID.Int64,
		Type:              in.Type.String,
		DefaultLanguageID: in.Strings[0].LanguageID.Int64,
		TeamsEditable:     *types.NewBool(false), // has no db default at the moment, so must be set
		NoScore:           *types.NewBool(false), // has no db default at the moment, so must be set
	}
}

func (in *NewItemRequest) groupItemData(id, userID, groupID int64) *database.GroupItem {
	return &database.GroupItem{
		ID:             *types.NewInt64(id),
		ItemID:         in.ID.Int64,
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
	return &database.ItemString{
		ID:          *types.NewInt64(id),
		ItemID:      in.ID.Int64,
		LanguageID:  in.Strings[0].LanguageID.Int64,
		Title:       in.Strings[0].Title.String,
		ImageURL:    in.Strings[0].ImageURL.String,
		Subtitle:    in.Strings[0].Subtitle.String,
		Description: in.Strings[0].Description.String,
	}
}
func (in *NewItemRequest) itemItemData(id int64) *database.ItemItem {
	return &database.ItemItem{
		ID:           *types.NewInt64(id),
		ChildItemID:  in.ID.Int64,
		Order:        in.Parents[0].Order.Int32,
		ParentItemID: in.Parents[0].ID.Int64,
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

	// check permissions
	if ret := srv.checkPermission(user, input.Parents[0].ID.Value.(int64)); ret != service.NoError {
		return ret
	}

	// insertion
	if err = srv.insertItem(user, input); err != nil {
		return service.ErrInvalidRequest(err)
	}

	// response
	response := struct {
		ItemID int64 `json:"ID,string"`
	}{input.ID.Value.(int64)}
	if err = render.Render(w, r, service.CreationSuccess(&response)); err != nil {
		return service.ErrUnexpected(err)
	}
	return service.NoError
}

func (srv *Service) insertItem(user *database.User, input *NewItemRequest) error {
	srv.Store.EnsureSetID(&input.ID.Int64)

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
