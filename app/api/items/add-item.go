package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	s "github.com/France-ioi/AlgoreaBackend/app/service"
	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID   t.OptionalInt64  `json:"id"`
	Type t.RequiredString `json:"type"`

	Strings []struct {
		LanguageID t.RequiredInt64  `json:"language_id"`
		Title      t.RequiredString `json:"title"`
	} `json:"strings"`

	Parents []struct {
		ID    t.RequiredInt64 `json:"id"`
		Order t.RequiredInt64 `json:"order"`
	} `json:"parents"`
}

// Bind validates the request body attributes
func (in *NewItemRequest) Bind(r *http.Request) error {
	if len(in.Strings) != 1 {
		return errors.New("Only one string per item is supported at the moment")
	}
	if len(in.Parents) != 1 {
		return errors.New("Only one parent item is supported at the moment")
	}
	return t.Validate(&in.ID, &in.Type)
}

func (in *NewItemRequest) itemData() *database.Item {
	return &database.Item{
		ID:   in.ID.Int64,
		Type: in.Type.String,
	}
}

func (in *NewItemRequest) groupItemData(id int64) *database.GroupItem {
	return &database.GroupItem{
		ID:             *t.NewInt64(id),
		ItemID:         in.ID.Int64,
		GroupID:        *t.NewInt64(6),        // dummy
		FullAccessDate: "2018-01-01 00:00:00", // dummy
	}
}

func (in *NewItemRequest) stringData(id int64) *database.ItemString {
	return &database.ItemString{
		ID:         *t.NewInt64(id),
		ItemID:     in.ID.Int64,
		LanguageID: in.Strings[0].LanguageID.Int64,
		Title:      in.Strings[0].Title.String,
	}
}
func (in *NewItemRequest) itemItemData(id int64) *database.ItemItem {
	return &database.ItemItem{
		ID:           *t.NewInt64(id),
		ChildItemID:  in.ID.Int64,
		Order:        in.Parents[0].Order.Int64,
		ParentItemID: in.Parents[0].ID.Int64,
	}
}

type Response struct {
	ItemID int64 `json:"ID"`
}

// ShowAccount godoc
// @Summary Show a account
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "Account ID"
// @Success 200 {object} items.Response
// @Router /accounts/{id} [get]
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
	resp := Response{input.ID.Value}
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
