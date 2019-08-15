package contests

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model contestGetGroupByNameResult
type getGroupByNameResult struct {
	// required: true
	GroupID int64 `gorm:"column:idGroup" json:"group_id,string"`
	// required: true
	AdditionalTime int32 `gorm:"column:iAdditionalTime" json:"additional_time"`
	// required: true
	TotalAdditionalTime int32 `gorm:"column:iTotalAdditionalTime" json:"total_additional_time"`
}

// swagger:operation GET /contests/{item_id}/group-by-name contests groups contestGetGroupByName
// ---
// summary: Get a group by name
// description: >
//                Return one group matching the name and satisfying:
//
//                  * the group has access to the contest (grayed, partial or full);
//                  * the authenticated user is an owner of the group;
//                  * the `groups.sName` (matching `sLogin` if a "UserSelf" group) is matching the input `name` parameter (case-insensitive)
//
//                If there are several groups or users matching, returns the first one (by `ID`).
//
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * the authenticated user should have `bCachedAccessSolutions` or `bCachedFullAccess` on the input item.
//
//                Otherwise, the "Forbidden" response is returned.
//
//
//                __NOTE__: This service is only here for transition between the former interface and the new one.
//                      This way of searching only by `sName`/`sLogin` and getting one result is not really convenient,
//                      but matching the former UI. This service will have to be removed as soon as
//                      the new interface is used.
//
// parameters:
// - name: item_id
//   description: "`ID` of a timed contest"
//   in: path
//   type: integer
//   required: true
// - name: name
//   in: query
//   type: string
//   required: true
// responses:
//   "200":
//     description: OK. Success response with the `group_id`, `additional_time`, `total_additional_time`
//     schema:
//       "$ref": "#/definitions/contestGetGroupByNameResult"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupByName(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupName, err := service.ResolveURLQueryGetStringField(r, "name")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	ok, err := srv.Store.Items().ByID(itemID).Where("items.sDuration IS NOT NULL").
		Joins("JOIN groups_items ON groups_items.idItem = items.ID").
		Joins(`
			JOIN groups_ancestors ON groups_ancestors.idGroupAncestor = groups_items.idGroup AND
				groups_ancestors.idGroupChild = ?`, user.SelfGroupID).
		Group("items.ID").
		Having("MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedAccessSolutionsDate) <= NOW()").
		HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.InsufficientAccessRightsError
	}

	var result getGroupByNameResult
	if err = srv.Store.Groups().OwnedBy(user).Where("groups.sName LIKE ?", groupName).
		Joins("JOIN groups_ancestors AS found_group_ancestors ON found_group_ancestors.idGroupChild = groups.ID").
		Joins("LEFT JOIN groups_items ON groups_items.idGroup = found_group_ancestors.idGroupAncestor AND groups_items.idItem = ?", itemID).
		Joins("LEFT JOIN groups_items AS main_group_item ON main_group_item.idGroup = groups.ID AND main_group_item.idItem = ?", itemID).
		Group("groups.ID").
		Having(`
			MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedPartialAccessDate) <= NOW() OR
			MIN(groups_items.sCachedGrayedAccessDate) <= NOW()`).
		Order("groups.ID").
		Select(`
			groups.ID AS idGroup,
			IFNULL(TIME_TO_SEC(main_group_item.sAdditionalTime), 0) AS iAdditionalTime,
			IFNULL(SUM(TIME_TO_SEC(groups_items.sAdditionalTime)), 0) AS iTotalAdditionalTime`).
		Take(&result).Error(); gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, result)
	return service.NoError
}
