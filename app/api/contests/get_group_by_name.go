package contests

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

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
//                If the contest is a team-only contest (`teams.bHasAttempts` is true) and the name matches an end-user,
//                returns his team instead of user’s 'selfgroup'.
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
//       "$ref": "#/definitions/contestInfo"
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

	isTeamOnly, err := srv.getTeamModeForTimedContestManagedByUser(itemID, user)
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	query := srv.Store.Groups().OwnedBy(user).
		Joins("JOIN groups_ancestors AS found_group_ancestors ON found_group_ancestors.idGroupChild = groups.ID").
		Joins("LEFT JOIN groups_items ON groups_items.idGroup = found_group_ancestors.idGroupAncestor AND groups_items.idItem = ?", itemID).
		Joins("LEFT JOIN groups_items AS main_group_item ON main_group_item.idGroup = groups.ID AND main_group_item.idItem = ?", itemID).
		Select(`
				groups.ID AS idGroup,
				groups.sName,
				groups.sType,
				IFNULL(TIME_TO_SEC(main_group_item.sAdditionalTime), 0) AS iAdditionalTime,
				IFNULL(SUM(TIME_TO_SEC(groups_items.sAdditionalTime)), 0) AS iTotalAdditionalTime`).
		Group("groups.ID").
		Having(`
			MIN(groups_items.sCachedFullAccessDate) <= NOW() OR MIN(groups_items.sCachedPartialAccessDate) <= NOW() OR
			MIN(groups_items.sCachedGrayedAccessDate) <= NOW()`).
		Order("groups.ID")

	if isTeamOnly {
		query = query.
			Joins(`
				LEFT JOIN groups_ancestors AS found_group_descendants
					ON found_group_descendants.idGroupAncestor = groups.ID`).
			Joins(`
				LEFT JOIN groups AS team
					ON team.ID = found_group_descendants.idGroupChild AND team.sType = 'Team' AND
						(groups.idTeamItem IN (SELECT idItemAncestor FROM items_ancestors WHERE idItemChild = ?) OR
						 groups.idTeamItem = ?)`, itemID, itemID).
			Joins(`
				LEFT JOIN groups_groups
					ON groups_groups.sType IN ('requestAccepted', 'invitationAccepted') AND
						groups_groups.idGroupParent = team.ID`).
			Joins(`
				LEFT JOIN groups AS user_group
					ON user_group.ID = groups_groups.idGroupChild AND user_group.sType = 'UserSelf' AND
						user_group.sName LIKE ?`, groupName).
			Group("groups.ID, user_group.ID").
			Having("MAX(user_group.ID) IS NOT NULL OR groups.sName LIKE ?", groupName)
	} else {
		query = query.
			Where("groups.sName LIKE ?", groupName)
	}

	var result contestInfo
	if err = query.Take(&result).Error(); gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, result)
	return service.NoError
}
