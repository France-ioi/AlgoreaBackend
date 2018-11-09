package items

import (
	"net/http"

	"github.com/go-chi/render"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
)

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID      int             `json:"id"`
	Type    string          `json:"type"`
	Strings []NewItemString `json:"strings"`
	Parents []NewItemParent `json:"parents"`
}

// NewItemString is a string record for new items
type NewItemString struct {
	LanguageID int    `json:"language_id"`
	Title      string `json:"title"`
}

// NewItemParent defines the parent items of a new item
type NewItemParent struct {
	ID    int `json:"id"`
	Order int `json:"order"`
}

func (*NewItemRequest) Bind(r *http.Request) error {
	// no more check/validation/update
	return nil
}

type newItemResponse struct {
}

func (srv *ItemService) addItem(w http.ResponseWriter, r *http.Request) {
	data := &NewItemRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, s.ErrInvalidRequest(err))
		return
	}
	if err := srv.Store.Items.Create(data.ID, data.Type, data.Strings[0].LanguageID, data.Strings[0].Title, data.Parents[0].ID, data.Parents[0].Order); err != nil {
		render.Render(w, r, s.ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, &newItemResponse{})
}
