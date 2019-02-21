package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getList(w http.ResponseWriter, r *http.Request) service.APIError {
	// Get IDs from request and validate it.
	ids, err := idsFromRequest(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	// Validate that the user can see the item IDs.
	user := srv.GetUser(r)
	if valid, err := srv.Store.Items().ValidateUserAccess(user, ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrForbidden(errors.New("insufficient access rights on given item ids"))
	}

	// Validate the hierarchy
	if valid, err := srv.Store.Items().IsValidHierarchy(ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrInvalidRequest(errors.New("the IDs chain is corrupt"))
	}

	// Build response
	// Fetch the requested items
	var items []struct {
		ItemID   int64  `json:"item_id"     sql:"column:idItem"`
		Title    string `json:"title"       sql:"column:sTitle"`
		Language int64  `json:"language_id" sql:"column:idLanguage"`
	}
	db := srv.Store.ItemStrings().Where("idItem IN (?)", ids).Scan(&items)
	if db.Error != nil {
		return service.ErrUnexpected(db.Error)
	}

	render.Respond(w, r, items)
	return service.NoError
}

func idsFromRequest(r *http.Request) ([]int64, error) {
	ids, err := service.QueryParamToInt64Slice(r, "ids")
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("no ids given")
	}
	if len(ids) > 10 {
		return nil, errors.New("no more than 10 ids expected")
	}
	return ids, nil
}
