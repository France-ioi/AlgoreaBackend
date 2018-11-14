package items

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"

	"github.com/go-chi/render"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID      t.OptionalInt64  `json:"id"`
	Type    t.RequiredString `json:"type"`
	Strings []NewItemString  `json:"strings"`
	Parents []NewItemParent  `json:"parents"`
}

// NewItemString is a string record for new items
type NewItemString struct {
	LanguageID t.RequiredInt64  `json:"language_id"`
	Title      t.RequiredString `json:"title"`
}

// NewItemParent defines the parent items of a new item
type NewItemParent struct {
	ID    t.RequiredInt64 `json:"id"`
	Order t.RequiredInt64 `json:"order"`
}

// Bind validates the request body attributes
func (i *NewItemRequest) Bind(r *http.Request) error {
	return t.Validate(&i.ID, &i.Type)
}

func (i *NewItemRequest) itemData() *database.Item {
	return &database.Item{
		ID:   i.ID.Int64,
		Type: i.Type.String,
	}
}

type newItemResponse struct {
}

func (srv *ItemService) addItem(w http.ResponseWriter, r *http.Request) {
	data := &NewItemRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, s.ErrInvalidRequest(err))
		return
	}
	if err := srv.Store.Items.Create(data.itemData(), data.Strings[0].LanguageID.Int64, data.Strings[0].Title.String, data.Parents[0].ID.Int64, data.Parents[0].Order.Int64); err != nil {
		render.Render(w, r, s.ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, &newItemResponse{})
}
