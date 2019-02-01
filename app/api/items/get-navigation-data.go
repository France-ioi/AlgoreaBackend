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

// GetItemRequest .
type GetItemRequest struct {
	ID int64 `json:"id"`
}

// Bind .
func (req *GetItemRequest) Bind(r *http.Request) error {
	strItemID := chi.URLParam(r, "itemID")
	itemID, err := strconv.ParseInt(strItemID, 10, 64)
	if err != nil {
		return fmt.Errorf("missing itemID")
	}
	req.ID = itemID
	return nil
}

func (srv *Service) getNavigationData(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	// Validate that the user has access to the root element.
	user := srv.getUser(httpReq)
	if valid, err := srv.Store.Items().ValidateUserAccess(user, []int64{req.ID}); err != nil {
		return service.ErrUnexpected(err)
	} else if !valid {
		return service.ErrForbidden(errors.New("insufficient access on given item ids"))
	}

	var defaultLanguageID int64 = 1
	result, err := srv.Store.Items().GetRawNavigationData(req.ID, user.UserID, user.DefaultLanguageID(), defaultLanguageID)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	// TODO:
	//	filter by the user's access rights
	//  construct the tree,
	//  filter the data fields,
	//  use a separate structure for the response
	render.Respond(rw, httpReq, result)
	return service.NoError
	/*
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
	*/
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
