package contests

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /contests/{item_id}/groups/{group_id}/members/additional-times contests groups contestListMembersAdditionalTime
// ---
// summary: Get additional times for a group of users/teams on a contest
// description: >
//                For all
//
//                  * descendant teams linked to the item via `idTeamItem` if `items.bHasAttempts`
//                  * end-users groups otherwise
//
//                having at least grayed access to the item, the service returns their
//                `group_id`, `name`, `type` and `additional_time` & `total_additional_time`.
//
//
//                * `additional_time` defaults to 0 if no such `groups_items`
//
//                * `total_additional_time` is the sum of additional times of this group on the item through all its
//                  `groups_ancestors` (even from different branches, but each ancestors counted only once), defaulting to 0
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * the authenticated user should have `solutions` or `full` access on the input item;
//                  * the authenticated user should own the `group_id`.
// parameters:
// - name: item_id
//   description: "`ID` of a timed contest"
//   in: path
//   type: integer
//   required: true
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: from.name
//   description: Start the page from the group next to the group with `name` = `from.name` and `ID` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the group next to the group with `name` = `from.name` and `ID`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [name,id]
//   type: array
//   items:
//     type: string
//     enum: [name,-name,id,-id]
// - name: limit
//   description: Display the first N groups
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with contests info
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/contestInfo"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getMembersAdditionalTimes(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	isTeamOnly, err := srv.getTeamModeForTimedContestManagedByUser(itemID, user)
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	ok, err := srv.Store.Groups().OwnedBy(user).Where("groups.ID = ?", groupID).HasRows()
	service.MustNotBeError(err)
	if !ok {
		return service.InsufficientAccessRightsError
	}

	query := srv.Store.GroupAncestors().Where("groups_ancestors.idGroupAncestor = ?", groupID)

	if isTeamOnly {
		query = query.
			Joins(`
				JOIN `+"`groups`"+` AS found_group
					ON found_group.ID = groups_ancestors.idGroupChild AND found_group.sType = 'Team' AND
						(found_group.idTeamItem IN (SELECT idItemAncestor FROM items_ancestors WHERE idItemChild = ?) OR
						 found_group.idTeamItem = ?)`, itemID, itemID)
	} else {
		query = query.
			Joins(`
				JOIN ` + "`groups`" + ` AS found_group
					ON found_group.ID = groups_ancestors.idGroupChild AND found_group.sType = 'UserSelf'`)
	}

	query = query.
		Joins("JOIN groups_ancestors AS found_group_ancestors ON found_group_ancestors.idGroupChild = found_group.ID").
		Joins("LEFT JOIN groups_items ON groups_items.idGroup = found_group_ancestors.idGroupAncestor AND groups_items.idItem = ?", itemID).
		Joins("LEFT JOIN groups_items AS main_group_item ON main_group_item.idGroup = found_group.ID AND main_group_item.idItem = ?", itemID).
		Select(`
				found_group.ID AS idGroup,
				found_group.sName,
				found_group.sType,
				IFNULL(TIME_TO_SEC(MAX(main_group_item.sAdditionalTime)), 0) AS iAdditionalTime,
				IFNULL(SUM(TIME_TO_SEC(groups_items.sAdditionalTime)), 0) AS iTotalAdditionalTime`).
		Group("found_group.ID").
		Having(`
			MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedPartialAccessDate) <= NOW() OR
			MIN(groups_items.sCachedGrayedAccessDate) <= NOW()`)

	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name": {ColumnName: "found_group.sName", FieldType: "string"},
			"id":   {ColumnName: "found_group.ID", FieldType: "int64"}},
		"name,id")
	if apiError != service.NoError {
		return apiError
	}

	var result []contestInfo
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
