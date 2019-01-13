package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
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
	// TODO: read language from somewhere
	languageID := int64(1)

	// Fetch information about the root item.
	dbItem, err := srv.Store.Items().GetOne(req.ID, languageID)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	item := treeItemFromDB(dbItem)
	if err := srv.buildChildrenStructure(item, languageID); err != nil {
		return service.ErrUnexpected(err)
	}

	render.Respond(rw, httpReq, item)
	return service.NoError
}

func (srv *Service) buildChildrenStructure(item *Item, languageID int64) error {
	allChildren, err := srv.Store.Items().GetChildrenOf(item.ItemID, languageID)
	if err != nil {
		return err
	}

	directChildren := childrenOf(item.ItemID, allChildren)
	item.fillChildren(directChildren)

	for i, ch := range item.Children {
		grandChildren := childrenOf(ch.ItemID, allChildren)
		item.Children[i].fillChildren(grandChildren)
	}

	return nil
}

func childrenOf(parentID int64, items []*database.TreeItem) []*database.TreeItem {
	var children []*database.TreeItem
	for _, it := range items {
		if it.ParentID == parentID {
			children = append(children, it)
		}
	}
	return children
}
