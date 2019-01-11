package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID   types.OptionalInt64  `json:"id"`
	Type types.RequiredString `json:"type"`

	Strings []struct {
		LanguageID types.RequiredInt64  `json:"language_id"`
		Title      types.RequiredString `json:"title"`
	} `json:"strings"`

	Parents []struct {
		ID    types.RequiredInt64 `json:"id"`
		Order types.RequiredInt64 `json:"order"`
	} `json:"parents"`
}

// Bind validates the request body attributes
func (in *NewItemRequest) Bind(r *http.Request) error {
	if len(in.Strings) != 1 {
		return errors.New("Exactly one string per item is supported at the moment")
	}
	if len(in.Parents) != 1 {
		return errors.New("Exactly one parent item is supported at the moment")
	}
	return types.Validate(&in.ID, &in.Type)
}

func (in *NewItemRequest) itemData() *database.Item {
	return &database.Item{
		ID:   in.ID.Int64,
		Type: in.Type.String,
	}
}

func (in *NewItemRequest) groupItemData(id int64, userID int64, groupID int64) *database.GroupItem {
	return &database.GroupItem{
		ID:             *types.NewInt64(id),
		ItemID:         in.ID.Int64,
		GroupID:        *types.NewInt64(groupID),
		CreatorUserID:  *types.NewInt64(userID),
		FullAccessDate: *types.NewDatetime(time.Now()),
		OwnerAccess:    *types.NewBool(true),
	}
}

func (in *NewItemRequest) stringData(id int64) *database.ItemString {
	return &database.ItemString{
		ID:         *types.NewInt64(id),
		ItemID:     in.ID.Int64,
		LanguageID: in.Strings[0].LanguageID.Int64,
		Title:      in.Strings[0].Title.String,
	}
}
func (in *NewItemRequest) itemItemData(id int64) *database.ItemItem {
	return &database.ItemItem{
		ID:           *types.NewInt64(id),
		ChildItemID:  in.ID.Int64,
		Order:        in.Parents[0].Order.Int64,
		ParentItemID: in.Parents[0].ID.Int64,
	}
}

func (srv *Service) addItem(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.getUser(r)

	// validate input (could be moved to JSON validation later)
	input := &NewItemRequest{}
	if err = render.Bind(r, input); err != nil {
		return service.ErrInvalidRequest(err)
	}

	// check permissions
	if ret := srv.checkPermission(srv.getUser(r), input.Parents[0].ID.Value); ret != service.NoError {
		return ret
	}

	// insertion
	if err = srv.insertItem(user, input); err != nil {
		return service.ErrInvalidRequest(err)
	}

	// response
	response := struct {
		ItemID int64 `json:"ID"`
	}{input.ID.Value}
	if err = render.Render(w, r, service.CreationSuccess(&response)); err != nil {
		return service.ErrUnexpected(err)
	}
	return service.NoError
}

func (srv *Service) insertItem(user *auth.User, input *NewItemRequest) error {
	srv.Store.EnsureSetID(&input.ID.Int64)

	return srv.Store.InTransaction(func(store *database.DataStore) error {
		var err error
		if err = store.Items().Insert(input.itemData()); err != nil {
			return err
		}
		if err = store.GroupItems().Insert(input.groupItemData(store.NewID(), user.UserID, user.SelfGroupID())); err != nil {
			return err
		}
		if err = store.ItemStrings().Insert(input.stringData(store.NewID())); err != nil {
			return err
		}
		return store.ItemItems().Insert(input.itemItemData(store.NewID()))
	})
}

func (srv *Service) checkPermission(user *auth.User, parentItemID int64) service.APIError {
	// can add a parent only if manager of that parent
	found, hasAccess, err := srv.Store.Items().HasManagerAccess(user, parentItemID)
	if err != nil {
		return service.ErrUnexpected(err)
	}
	if !found {
		return service.ErrForbidden(errors.New("Cannot find the parent item"))
	}
	if !hasAccess {
		return service.ErrForbidden(errors.New("Insufficient access on the parent item"))
	}
	return service.NoError
}
