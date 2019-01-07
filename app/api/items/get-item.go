package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type GetItemRequest struct {
	ID int64 `json:"id"`
}

func (req *GetItemRequest) Bind(r *http.Request) error {
	strItemID := chi.URLParam(r, "itemID")
	itemID, err := strconv.Atoi(strItemID)
	if err != nil {
		return fmt.Errorf("missing itemID")
	}
	req.ID = int64(itemID)
	return nil
}

func (srv *Service) getItem(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	// Validate that the user has access to the root element.
	user := srv.getUser(httpReq)
	if valid, err := srv.Store.Items().ValidateUserAccess(user, []int64{req.ID}); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrForbidden(errors.New("Insufficient access on given item ids"))
	}

	// Fetch information about the root item.
	dbItem, err := srv.Store.Items().Get(req.ID)
	if err != nil {
		return service.ErrUnexpected(err)
	}
	dbItemString, err := srv.Store.ItemStrings().GetByItemID(req.ID)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	item := &Item{}
	item.fillItemData(dbItem)
	item.fillItemStringData(dbItemString)
	if err := srv.buildChildrenStructure(item); err != nil {
		return service.ErrUnexpected(err)
	}
	for i := range item.Children {
		if err := srv.buildChildrenStructure(item.Children[i]); err != nil {
			return service.ErrUnexpected(err)
		}
	}

	render.Respond(rw, httpReq, item)
	return service.NoError
}

func (srv *Service) buildChildrenStructure(item *Item) error {
	// Fetch information about the children items.
	dbChildrenItemItems, err := srv.Store.ItemItems().ChildrenOf(item.ItemID)
	if err != nil {
		return err
	}
	childrenIDs := make([]int64, 0, len(dbChildrenItemItems))
	for _, chIt := range dbChildrenItemItems {
		childrenIDs = append(childrenIDs, chIt.ChildItemID.Value)
	}
	dbChildrenItems, err := srv.Store.Items().ListByIDs(childrenIDs)
	if err != nil {
		return err
	}
	dbChildrenItemStrings, err := srv.Store.ItemStrings().GetByItemIDs(childrenIDs)
	if err != nil {
		return err
	}
	item.fillChildren(dbChildrenItems, dbChildrenItemItems, dbChildrenItemStrings)
	return nil
}
