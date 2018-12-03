package items

import (
  "fmt"
	"net/http"

	s "github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/render"
)

// ListResponseRow is the structure of a row for the service response
type ListResponseRow struct {
	ItemID    int64  `json:"item_id"     sql:"column:idItem"`
	Title     string `json:"title"       sql:"column:sTitle"`
	Language  int64  `json:"language_id" sql:"column:idLanguage"`
}

func (srv *Service) getList(w http.ResponseWriter, r *http.Request) s.APIError {
  var err error

  // Validate the input data
	var ids []int64
	if ids, err = s.QueryParamToInt64Slice(r, "ids"); err != nil {
		return s.ErrInvalidRequest(err)
	}
	if len(ids) == 0 {
		return s.ErrInvalidRequest(fmt.Errorf("No ids given"))
	}
	if len(ids) > 10 {
		return s.ErrInvalidRequest(fmt.Errorf("Maximum ids expected"))
	}

  // Validate that the user can see the item ids
  var valid bool
  if valid, err = srv.Store.Items.ValidateUserAccess(ids); err != nil {
    return s.ErrUnexpected(err)
  }
  if !valid {
    return s.ErrForbidden(fmt.Errorf("Insufficient access on given item ids"))
  }

  // Todo: validate the hierarchy
  // srv.Store.Items.IsValidHierarchy(...)

  // Fetch the requested items
  items := []ListResponseRow{}
	err = srv.Store.Items.GetList(ids, &items)

	if err != nil {
		return s.ErrUnexpected(err)
	}
	render.Respond(w, r, items)
	return s.NoError
}
