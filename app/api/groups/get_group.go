package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id} groups groupView
// ---
// summary: Get group info
// description: >
//
//   Returns general information about the group from the `groups` table.
//
//
//   The authenticated user should be an owner of `group_id` OR a descendant of the group OR  the group's `free_access`=1,
//   otherwise the 'forbidden' error is returned.
//
//
//   Note: `code*` fields are omitted when the user is not an owner of the group.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     description: OK. The group info
//     schema:
//       type: object
//       properties:
//         id:
//           description: group's `ID`
//           type: string
//           format: int64
//         name:
//           type: string
//         grade:
//           type: integer
//         description:
//           description: Nullable
//           type: string
//         date_created:
//           description: Nullable
//           type: string
//         type:
//           type: string
//           enum: [Class,Team,Club,Friends,Other,UserSelf]
//         redirect_path:
//           description: Nullable
//           type: string
//         opened:
//           type: boolean
//         free_access:
//           type: boolean
//         code:
//           description: Nullable
//           type: string
//         code_timer:
//           description: Nullable
//           type: string
//         code_end:
//           description: Nullable
//           type: string
//         open_contest:
//           type: boolean
//         current_user_is_owner:
//           type: boolean
//         current_user_is_member:
//           description: >
//                          `True` when there is an active group->user relation in `groups_groups`
//           type: boolean
//       required: [id, name, grade, description, date_created, type, redirect_path, opened, free_access,
//                  open_contest, current_user_is_owner, current_user_is_member]
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	query := srv.Store.Groups().
		Joins(`
			LEFT JOIN groups_ancestors
				ON groups_ancestors.group_child_id = groups.id AND groups_ancestors.group_ancestor_id = ?`, user.OwnedGroupID).
		Joins(`
			LEFT JOIN groups_ancestors AS groups_descendants
				ON groups_descendants.group_ancestor_id = groups.id AND groups_descendants.group_child_id = ?`, user.SelfGroupID).
		Joins(`
			LEFT JOIN groups_groups
				ON groups_groups.type `+database.GroupRelationIsActiveCondition+` AND
					groups_groups.group_parent_id = groups.id AND groups_groups.group_child_id = ?`, user.SelfGroupID).
		Where("groups_ancestors.id IS NOT NULL OR groups_descendants.id IS NOT NULL OR groups.free_access").
		Where("groups.id = ?", groupID).Select(
		`groups.id, groups.name, groups.grade, groups.description, groups.date_created,
			groups.type, groups.redirect_path, groups.opened, groups.free_access,
			IF(groups_ancestors.id IS NOT NULL, groups.code, NULL) AS code,
			IF(groups_ancestors.id IS NOT NULL, groups.code_timer, NULL) AS code_timer,
			IF(groups_ancestors.id IS NOT NULL, groups.code_end, NULL) AS code_end,
			groups.open_contest,
			groups_ancestors.id IS NOT NULL AS current_user_is_owner,
			groups_groups.id IS NOT NULL AS current_user_is_member`).Limit(1)

	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}

	jsonResult := service.ConvertMapFromDBToJSON(result[0])
	if !jsonResult["current_user_is_owner"].(bool) {
		delete(jsonResult, "code")
		delete(jsonResult, "code_timer")
		delete(jsonResult, "code_end")
	}
	for _, key := range [...]string{"code_end", "date_created"} {
		if value, ok := jsonResult[key]; ok && value != nil {
			parsedTime := &database.Time{}
			service.MustNotBeError(parsedTime.ScanString(value.(string)))
			jsonResult[key] = parsedTime
		}
	}

	render.Respond(w, r, jsonResult)

	return service.NoError
}
