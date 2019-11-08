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

type navigationItemUserActiveAttempt struct {
	Score       float32        `json:"score"`
	Validated   bool           `json:"validated"`
	Finished    bool           `json:"finished"`
	KeyObtained bool           `json:"key_obtained"`
	Submissions int32          `json:"submissions"`
	StartedAt   *database.Time `json:"started_at"`
	ValidatedAt *database.Time `json:"validated_at"`
	FinishedAt  *database.Time `json:"finished_at"`
}

type navigationItemAccessRights struct {
	CanView string `json:"can_view"`
}

type navigationItemString struct {
	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title *string `json:"title"`
}

type navigationItemCommonFields struct {
	ID   int64  `json:"id,string"`
	Type string `json:"type"`
	// whether items.unlocked_item_ids is empty
	HasUnlockedItems bool `json:"has_unlocked_items"`

	String            navigationItemString             `json:"string"`
	UserActiveAttempt *navigationItemUserActiveAttempt `json:"user_active_attempt"`
	AccessRights      navigationItemAccessRights       `json:"access_rights"`

	Children []navigationItemChild `json:"children"`
}

type navigationDataResponse struct {
	*navigationItemCommonFields
}

type navigationItemChild struct {
	*navigationItemCommonFields

	Order                  int32  `json:"order"`
	ContentViewPropagation string `json:"content_view_propagation"`
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
		if !hasParentItem || parentItem.CanViewGeneratedValue == srv.Store.PermissionsGranted().ViewIndexByName("info") {
			continue // Only 'info' access to the parent item
		}

		if parentItemCommonFields, ok := idsToResponseData[rawData[index].ParentItemID]; ok {
			child := navigationItemChild{
				navigationItemCommonFields: srv.fillNavigationCommonFieldsWithDBData(&rawData[index]),
				Order:                      rawData[index].Order,
				ContentViewPropagation:     rawData[index].ContentViewPropagation,
			}
			idsToResponseData[child.ID] = child.navigationItemCommonFields
			parentItemCommonFields.Children = append(parentItemCommonFields.Children, child)
		}
	}
}

func (srv *Service) fillNavigationCommonFieldsWithDBData(rawData *rawNavigationItem) *navigationItemCommonFields {
	result := &navigationItemCommonFields{
		ID:               rawData.ID,
		Type:             rawData.Type,
		HasUnlockedItems: rawData.HasUnlockedItems,
		String:           navigationItemString{Title: rawData.Title},
		AccessRights: navigationItemAccessRights{
			CanView: srv.Store.PermissionsGranted().ViewNameByIndex(rawData.CanViewGeneratedValue),
		},
	}
	if rawData.ItemGrandparentID == nil {
		result.Children = make([]navigationItemChild, 0)
	}
	if rawData.UserAttemptID != nil {
		result.UserActiveAttempt = &navigationItemUserActiveAttempt{
			Score:       rawData.UserScore,
			Validated:   rawData.UserValidated,
			Finished:    rawData.UserFinished,
			KeyObtained: rawData.UserKeyObtained,
			Submissions: rawData.UserSubmissions,
			StartedAt:   rawData.UserStartedAt,
			ValidatedAt: rawData.UserValidatedAt,
			FinishedAt:  rawData.UserFinishedAt,
		}
	}
	return result
}
