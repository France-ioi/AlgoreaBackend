package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type navigationItemAccessRights struct {
	// required: true
	// enum: none,info,content,content_with_descendants,solution
	CanView string `json:"can_view"`
}

type navigationItemString struct {
	// [Nullable] Title (from `items_strings`) in the user’s default language or (if not available) default language of the item
	// required: true
	Title *string `json:"title"`
}

type navigationItemCommonFields struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type"`

	// required: true
	String navigationItemString `json:"string"`

	// max among all attempts of the user (or of the team given in `as_team_id`)
	// required: true
	BestScore float32 `json:"best_score"`
	// max among all attempts of the user (or of the team given in `as_team_id`)
	// required: true
	Validated bool `json:"validated"`

	// required: true
	AccessRights navigationItemAccessRights `json:"access_rights"`

	// Nullable
	// required: true
	Children []navigationItemChild `json:"children"`
}

// swagger:model itemNavTreeResponse
type navTreeResponse struct {
	*navigationItemCommonFields
}

type navigationItemChild struct {
	*navigationItemCommonFields

	// `items_items.child_order`
	// required: true
	Order int32 `json:"order"`
	// from `items_items`
	// required: true
	// enum: none,as_info,as_content
	ContentViewPropagation string `json:"content_view_propagation"`
}

// swagger:operation GET /items/{item_id}/nav-tree items itemNavTreeGet
// ---
// summary: Get the navigation tree of an item
// description: >
//
//   Returns data needed to display the navigation menu (for `item_id`, its children, and its grandchildren), only items
//   visible to the current user (or to the `as_team_id` team) are shown.
//
//
//   * If the specified `item_id` doesn't exist or is not visible to the current user (or to the `as_team_id` team),
//     the 'forbidden' response is returned.
//
//
//   * If `as_team_id` is given, it should be a user's parent team group,
//     otherwise the "forbidden" error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
// responses:
//   "200":
//     description: OK. Navigation data
//     schema:
//       "$ref": "#/definitions/itemNavTreeResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItemNavTree(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	groupID, apiError := srv.getParticipantIDFromRequest(httpReq, user)
	if apiError != service.NoError {
		return apiError
	}

	rawData := getRawNavigationData(srv.Store, itemID, groupID, user)

	if len(rawData) == 0 || rawData[0].ID != itemID {
		return service.ErrForbidden(errors.New("insufficient access rights on given item id"))
	}

	response := navTreeResponse{
		srv.fillNavigationCommonFieldsWithDBData(&rawData[0]),
	}
	idMap := map[int64]*rawNavigationItem{}
	for index := range rawData {
		idMap[rawData[index].ID] = &rawData[index]
	}
	idsToResponseData := map[int64]*navigationItemCommonFields{itemID: response.navigationItemCommonFields}
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
		ID:        rawData.ID,
		Type:      rawData.Type,
		String:    navigationItemString{Title: rawData.Title},
		BestScore: rawData.UserBestScore,
		Validated: rawData.UserValidated,
		AccessRights: navigationItemAccessRights{
			CanView: srv.Store.PermissionsGranted().ViewNameByIndex(rawData.CanViewGeneratedValue),
		},
	}
	if rawData.ItemGrandparentID == nil {
		result.Children = make([]navigationItemChild, 0)
	}
	return result
}
