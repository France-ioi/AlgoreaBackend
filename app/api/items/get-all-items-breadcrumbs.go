package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getUser(r *http.Request) *auth.User {
	return auth.UserFromContext(r.Context(), srv.Store.Users())
}

func (srv *Service) getList(w http.ResponseWriter, r *http.Request) service.APIError {
	// Get IDs from request and validate it.
	ids, err := idsFromRequest(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	// Validate that the user can see the item IDs.
	user := srv.getUser(r)
	if valid, err := srv.Store.Items().ValidateUserAccess(user, ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrForbidden(errors.New("Insufficient access on given item ids"))
	}

	// Validate the hierarchy
	if valid, err := srv.Store.Items().IsValidHierarchy(ids); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrInvalidRequest(errors.New("The IDs chain is corrupt"))
	}

	// Build response
	// Fetch the requested items
	var items []struct {
		ItemID   int64  `json:"item_id"     sql:"column:idItem"`
		Title    string `json:"title"       sql:"column:sTitle"`
		Language int64  `json:"language_id" sql:"column:idLanguage"`
	}
	db := srv.Store.ItemStrings().All().Where("idItem IN (?)", ids).Scan(&items)
	if db.Error() != nil {
		return service.ErrUnexpected(db.Error())
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
		return nil, errors.New("No ids given")
	}
	if len(ids) > 10 {
		return nil, errors.New("Maximum ids expected")
	}
	return ids, nil
}
