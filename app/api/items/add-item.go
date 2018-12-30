package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	s "github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/types"
)

// Bind validates the request body attributes
func (in *NewItemRequest) Bind(r *http.Request) error {
	if len(in.Strings) != 1 {
		return errors.New("Only one string per item is supported at the moment")
	}
	if len(in.Parents) != 1 {
		return errors.New("Only one parent item is supported at the moment")
	}
	return types.Validate(&in.ID, &in.Type)
}

func (in *NewItemRequest) itemData() *database.Item {
	return &database.Item{
		ID:   in.ID.Int64,
		Type: in.Type.String,
	}
}

func (in *NewItemRequest) groupItemData(id int64) *database.GroupItem {
	return &database.GroupItem{
		ID:             *types.NewInt64(id),
		ItemID:         in.ID.Int64,
		GroupID:        *types.NewInt64(6),    // dummy
		FullAccessDate: "2018-01-01 00:00:00", // dummy
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

// ShowAccount godoc
// @Summary Add an item
// @Description Add an item within a hierarchy.
// @ID add-item
// @Accept  json
// @Produce  json
// @Param body body items.NewItemRequest true "{}"
// @Success 200 {object} items.NewItemResponse
// @Router /items/ [post]
// @Security cookieAuth
func (srv *Service) addItem(w http.ResponseWriter, r *http.Request) s.APIError {
	var err error

	// validate input (could be moved to JSON validation later)
	input := &NewItemRequest{}
	if err = render.Bind(r, input); err != nil {
		return s.ErrInvalidRequest(err)
	}

	// insertion
	if err = srv.insertItem(input); err != nil {
		return s.ErrInvalidRequest(err)
	}

	// response
	resp := NewItemResponse{input.ID.Value}
	if err = render.Render(w, r, s.CreationSuccess(&resp)); err != nil {
		return s.ErrUnexpected(err)
	}
	return s.NoError
}

func (srv *Service) insertItem(input *NewItemRequest) error {
	srv.Store.EnsureSetID(&input.ID.Int64)

	return srv.Store.InTransaction(func(store *database.DataStore) error {
		var err error
		if err = store.Items().Insert(input.itemData()); err != nil {
			return err
		}
		if err = store.GroupItems().Insert(input.groupItemData(store.NewID())); err != nil {
			return err
		}
		if err = store.ItemStrings().Insert(input.stringData(store.NewID())); err != nil {
			return err
		}
		return store.ItemItems().Insert(input.itemItemData(store.NewID()))
	})
}
