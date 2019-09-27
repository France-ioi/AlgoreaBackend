package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// GetItemRequest wraps the id parameter
type GetItemRequest struct {
	ID int64 `json:"id"`
}

type navigationItemUser struct {
	Score               float32        `json:"score"`
	Validated           bool           `json:"validated"`
	Finished            bool           `json:"finished"`
	KeyObtained         bool           `json:"key_obtained"`
	SubmissionsAttempts int32          `json:"submissions_attempts"`
	StartDate           *database.Time `json:"start_date"`
	ValidationDate      *database.Time `json:"validation_date"`
	FinishDate          *database.Time `json:"finish_date"`
}

type navigationItemAccessRights struct {
	FullAccess    bool `json:"full_access"`
	PartialAccess bool `json:"partial_access"`
	GrayAccess    bool `json:"gray_access"`
}

type navigationItemString struct {
	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title *string `json:"title"`
}

type navigationItemCommonFields struct {
	ID                int64  `json:"id,string"`
	Type              string `json:"type"`
	TransparentFolder bool   `json:"transparent_folder"`
	// whether items.unlocked_item_ids is empty
	HasUnlockedItems bool `json:"has_unlocked_items"`

	String       navigationItemString       `json:"string"`
	User         navigationItemUser         `json:"user"`
	AccessRights navigationItemAccessRights `json:"access_rights"`

	Children []navigationItemChild `json:"children"`
}

type navigationDataResponse struct {
	*navigationItemCommonFields
}

type navigationItemChild struct {
	*navigationItemCommonFields

	Order                    int32  `json:"order"`
	PartialAccessPropagation string `json:"partial_access_propagation"`
}

// Bind binds req.ID to URLParam("item_id")
func (req *GetItemRequest) Bind(r *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return err
	}
	req.ID = itemID
	return nil
}

func (srv *Service) getNavigationData(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	rawData, err := getRawNavigationData(srv.Store, req.ID, user)
	service.MustNotBeError(err)

	if len(rawData) == 0 || rawData[0].ID != req.ID {
		return service.ErrForbidden(errors.New("insufficient access rights on given item id"))
	}

	response := navigationDataResponse{
		srv.fillNavigationCommonFieldsWithDBData(&rawData[0]),
	}
	idMap := map[int64]*rawNavigationItem{}
	for index := range rawData {
		idMap[rawData[index].ID] = &rawData[index]
	}
	idsToResponseData := map[int64]*navigationItemCommonFields{req.ID: response.navigationItemCommonFields}
	srv.fillNavigationSubtreeWithChildren(rawData, idMap, idsToResponseData)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func (srv *Service) fillNavigationSubtreeWithChildren(rawData []rawNavigationItem,
	idMap map[int64]*rawNavigationItem,
	idsToResponseData map[int64]*navigationItemCommonFields) {
	for index := range rawData {
		if index == 0 {
			continue
		}

		parentItem, hasParentItem := idMap[rawData[index].ParentItemID]
		if !hasParentItem ||
			(!parentItem.FullAccess && !parentItem.PartialAccess) {
			continue // The parent item is grayed
		}

		if parentItemCommonFields, ok := idsToResponseData[rawData[index].ParentItemID]; ok {
			child := navigationItemChild{
				navigationItemCommonFields: srv.fillNavigationCommonFieldsWithDBData(&rawData[index]),
				Order:                      rawData[index].Order,
				PartialAccessPropagation:   rawData[index].PartialAccessPropagation,
			}
			idsToResponseData[child.ID] = child.navigationItemCommonFields
			parentItemCommonFields.Children = append(parentItemCommonFields.Children, child)
		}
	}
}

func (srv *Service) fillNavigationCommonFieldsWithDBData(rawData *rawNavigationItem) *navigationItemCommonFields {
	result := &navigationItemCommonFields{
		ID:                rawData.ID,
		Type:              rawData.Type,
		TransparentFolder: rawData.TransparentFolder,
		HasUnlockedItems:  rawData.HasUnlockedItems,
		String:            navigationItemString{Title: rawData.Title},
		User: navigationItemUser{
			Score:               rawData.UserScore,
			Validated:           rawData.UserValidated,
			Finished:            rawData.UserFinished,
			KeyObtained:         rawData.UserKeyObtained,
			SubmissionsAttempts: rawData.UserSubmissionsAttempts,
			StartDate:           rawData.UserStartDate,
			ValidationDate:      rawData.UserValidationDate,
			FinishDate:          rawData.UserFinishDate,
		},
		AccessRights: navigationItemAccessRights{
			FullAccess:    rawData.FullAccess,
			PartialAccess: rawData.PartialAccess,
			GrayAccess:    rawData.GrayedAccess,
		},
	}
	if rawData.ItemGrandparentID == nil {
		result.Children = make([]navigationItemChild, 0)
	}
	return result
}
