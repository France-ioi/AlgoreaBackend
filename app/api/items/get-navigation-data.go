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

type NavigationItemCommonFields struct {
	ID                		int64   `json:"item_id"`
	Type              		string  `json:"type"`
	TransparentFolder 		bool	  `json:"transparent_folder"`
	// whether items.idItemUnlocked is empty
	HasUnlockedItems  		bool    `json:"has_unlocked_items"`

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title         				string  `json:"title"`

	UserScore 						float32	`json:"user_score,omitempty"`
	UserValidated 				bool	  `json:"user_validated,omitempty"`
	UserFinished					bool	  `json:"user_finished,omitempty"`
	KeyObtained 					bool 	  `json:"key_obtained,omitempty"`
	SubmissionsAttempts   int64   `json:"submissions_attempts,omitempty"`
	StartDate             string  `json:"start_date,omitempty"` // iso8601 str
	ValidationDate        string  `json:"validation_date,omitempty"` // iso8601 str
	FinishDate            string  `json:"finish_date,omitempty"` // iso8601 str

	FullAccess						bool		`json:"full_access"`
	PartialAccess					bool		`json:"partial_access"`
	GrayAccess  					bool		`json:"gray_access"`

	Children							[]NavigationItemChild `json:"children,omitempty"`
}

type NavigationDataResponse struct {
	*NavigationItemCommonFields
}

type NavigationItemChild struct {
	*NavigationItemCommonFields

	Order 						int64 `json:"order"`
	AccessRestricted  bool  `json:"access_restricted"`
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

	user := srv.getUser(httpReq)
	var defaultLanguageID int64 = 1
	rawData, err := srv.Store.Items().GetRawNavigationData(req.ID, user.UserID, user.DefaultLanguageID(), defaultLanguageID)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	var ids []int64
	for _, row := range *rawData {
		ids = append(ids, row.ID)
	}
	accessDetailsMap, err := srv.Store.Items().GetAccessDetailsMapForIDs(user, ids)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	accessDetailsForRootItem, hasAccessDetailsForRootItem := accessDetailsMap[req.ID]
	if len(*rawData) == 0 || (*rawData)[0].ID != req.ID || !hasAccessDetailsForRootItem ||
		(!accessDetailsForRootItem.FullAccess && !accessDetailsForRootItem.PartialAccess && !accessDetailsForRootItem.GrayedAccess){
		return service.ErrForbidden(errors.New("insufficient access rights on given item id"))
	}

	response := NavigationDataResponse{
		srv.fillNavigationCommonFieldsWithDBData(&(*rawData)[0], &accessDetailsForRootItem),
	}
	idsToResponseData := map[int64]*NavigationItemCommonFields{req.ID: response.NavigationItemCommonFields}

	for index, item := range *rawData {
		if index == 0 {
			continue
		}

		accessDetailsForItem, hasAccessDetailsForItem := accessDetailsMap[item.ID]
		if !hasAccessDetailsForItem ||
			(!accessDetailsForItem.FullAccess && !accessDetailsForItem.PartialAccess &&
				!accessDetailsForItem.GrayedAccess) {
			continue // The user has no access to the item
		}

		accessDetailsForParentItem, hasAccessDetailsForParentItem := accessDetailsMap[item.IDItemParent]
		if !hasAccessDetailsForParentItem ||
			(!accessDetailsForParentItem.FullAccess && !accessDetailsForParentItem.PartialAccess) {
			continue // The parent item is grayed
		}

		if parentItemCommonFields, ok := idsToResponseData[item.IDItemParent]; ok {
			child := NavigationItemChild{
				NavigationItemCommonFields: srv.fillNavigationCommonFieldsWithDBData(&item, &accessDetailsForItem),
				Order: item.Order,
				AccessRestricted:	item.AccessRestricted,
			}
			idsToResponseData[child.ID] = child.NavigationItemCommonFields
			parentItemCommonFields.Children = append(parentItemCommonFields.Children, child)
		}
	}

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func (srv *Service) fillNavigationCommonFieldsWithDBData(
	  rawData *database.RawNavigationItem,
		accessDetail *database.ItemAccessDetails,
	)*NavigationItemCommonFields {
	return &NavigationItemCommonFields{
		ID: rawData.ID,
		Type: rawData.Type,
		TransparentFolder: rawData.TransparentFolder,
		HasUnlockedItems: rawData.HasUnlockedItems,
		Title: rawData.Title,
		UserScore: rawData.UserScore,
		UserValidated: rawData.UserValidated,
		UserFinished: rawData.UserFinished,
		KeyObtained: rawData.KeyObtained,
		SubmissionsAttempts: rawData.SubmissionsAttempts,
		StartDate: rawData.StartDate,
		ValidationDate: rawData.ValidationDate,
		FinishDate: rawData.FinishDate,

		FullAccess: accessDetail.FullAccess,
		PartialAccess: accessDetail.PartialAccess,
		GrayAccess: accessDetail.GrayedAccess,
	}
}
