package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// GetItemRequest .
type GetItemRequest struct {
	ID int64 `json:"id"`
}

type navigationItemUser struct {
	Score     						float32	`json:"score"`
	Validated     				bool	  `json:"validated"`
	Finished    					bool	  `json:"finished"`
	KeyObtained 					bool 	  `json:"key_obtained"`
	SubmissionsAttempts   int64   `json:"submissions_attempts"`
	StartDate             string  `json:"start_date"` // iso8601 str
	ValidationDate        string  `json:"validation_date"` // iso8601 str
	FinishDate            string  `json:"finish_date"` // iso8601 str
}

type navigationItemAccessRights struct {
	FullAccess						bool		`json:"full_access"`
	PartialAccess					bool		`json:"partial_access"`
	GrayAccess  					bool		`json:"gray_access"`
}

type navigationItemString struct {
	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title         				string  `json:"title"`
}

type navigationItemCommonFields struct {
	ID                		int64   `json:"id"`
	Type              		string  `json:"type"`
	TransparentFolder 		bool	  `json:"transparent_folder"`
	// whether items.idItemUnlocked is empty
	HasUnlockedItems  		bool    `json:"has_unlocked_items"`

	String                navigationItemString `json:"string"`
	User                  navigationItemUser `json:"user"`
	AccessRights          navigationItemAccessRights `json:"access_rights"`

	Children							[]navigationItemChild `json:"children,omitempty"`
}

type navigationDataResponse struct {
	*navigationItemCommonFields
}

type navigationItemChild struct {
	*navigationItemCommonFields

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
	rawData, err := srv.Store.Items().GetRawNavigationData(req.ID, user.UserID,
		user.DefaultLanguageID(), defaultLanguageID)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	accessDetailsMap, err := srv.getAccessDetailsForRawNavigationItems(rawData, user)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	accessDetailsForRootItem, hasAccessDetailsForRootItem := accessDetailsMap[req.ID]
	if len(*rawData) == 0 || (*rawData)[0].ID != req.ID || !hasAccessDetailsForRootItem ||
		(!accessDetailsForRootItem.FullAccess && !accessDetailsForRootItem.PartialAccess && !accessDetailsForRootItem.GrayedAccess) {
		return service.ErrForbidden(errors.New("insufficient access rights on given item id"))
	}

	response := navigationDataResponse{
		srv.fillNavigationCommonFieldsWithDBData(&(*rawData)[0], &accessDetailsForRootItem),
	}
	idsToResponseData := map[int64]*navigationItemCommonFields{req.ID: response.navigationItemCommonFields}
	srv.fillNavigationSubtreeWithChildren(rawData, accessDetailsMap, idsToResponseData)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func (srv *Service) fillNavigationSubtreeWithChildren(rawData *[]database.RawNavigationItem,
	accessDetailsMap map[int64]database.ItemAccessDetails,
	idsToResponseData map[int64]*navigationItemCommonFields) {
	for index, item := range *rawData {
		if index == 0 {
			continue
		}

		accessDetailsForItem, hasAccessDetailsForItem := accessDetailsMap[item.ID]
		if !hasAccessDetailsForItem || !hasSufficientAccessOnNavigationItem(accessDetailsForItem) {
			continue
		}

		accessDetailsForParentItem, hasAccessDetailsForParentItem := accessDetailsMap[item.IDItemParent]
		if !hasAccessDetailsForParentItem ||
			(!accessDetailsForParentItem.FullAccess && !accessDetailsForParentItem.PartialAccess) {
			continue // The parent item is grayed
		}

		if parentItemCommonFields, ok := idsToResponseData[item.IDItemParent]; ok {
			child := navigationItemChild{
				navigationItemCommonFields: srv.fillNavigationCommonFieldsWithDBData(&item, &accessDetailsForItem),
				Order:                      item.Order,
				AccessRestricted:           item.AccessRestricted,
			}
			idsToResponseData[child.ID] = child.navigationItemCommonFields
			parentItemCommonFields.Children = append(parentItemCommonFields.Children, child)
		}
	}
}

// hasSufficientAccessOnNavigationItem checks if the user has access rights on the item
func hasSufficientAccessOnNavigationItem(accessDetailsForItem database.ItemAccessDetails) bool {
	return accessDetailsForItem.FullAccess || accessDetailsForItem.PartialAccess ||
		accessDetailsForItem.GrayedAccess
}

func (srv *Service) getAccessDetailsForRawNavigationItems(rawData *[]database.RawNavigationItem, user *auth.User,
	) (map[int64]database.ItemAccessDetails, error) {
	var ids []int64
	for _, row := range *rawData {
		ids = append(ids, row.ID)
	}
	accessDetailsMap, err := srv.Store.Items().GetAccessDetailsMapForIDs(user, ids)
	return accessDetailsMap, err
}

func (srv *Service) fillNavigationCommonFieldsWithDBData(
	  rawData *database.RawNavigationItem,
		accessDetail *database.ItemAccessDetails,
	)*navigationItemCommonFields {
	return &navigationItemCommonFields{
		ID: rawData.ID,
		Type: rawData.Type,
		TransparentFolder: rawData.TransparentFolder,
		HasUnlockedItems: rawData.HasUnlockedItems,
		String: navigationItemString{ Title: rawData.Title },
		User: navigationItemUser{
			Score: rawData.UserScore,
			Validated: rawData.UserValidated,
			Finished: rawData.UserFinished,
			KeyObtained: rawData.UserKeyObtained,
			SubmissionsAttempts: rawData.UserSubmissionsAttempts,
			StartDate: rawData.UserStartDate,
			ValidationDate: rawData.UserValidationDate,
			FinishDate: rawData.UserFinishDate,
		},
		AccessRights: navigationItemAccessRights{
			FullAccess:    accessDetail.FullAccess,
			PartialAccess: accessDetail.PartialAccess,
			GrayAccess:    accessDetail.GrayedAccess,
		},
	}
}
