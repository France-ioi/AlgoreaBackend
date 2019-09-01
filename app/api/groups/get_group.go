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
//   The authenticated user should be an owner of `group_id` OR a descendant of the group OR  the group's `bFreeAccess`=1,
//   otherwise the 'forbidden' error is returned.
//
//
//   Note: `code_*` fields are nulls when the user is not an owner of the group.
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
//                  code, code_timer, code_end, open_contest, current_user_is_owner, current_user_is_member]
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
				ON groups_ancestors.idGroupChild = groups.ID AND groups_ancestors.idGroupAncestor = ?`, user.OwnedGroupID).
		Joins(`
			LEFT JOIN groups_ancestors AS groups_descendants
				ON groups_descendants.idGroupAncestor = groups.ID AND groups_descendants.idGroupChild = ?`, user.SelfGroupID).
		Joins(`
			LEFT JOIN groups_groups
				ON groups_groups.sType `+database.GroupRelationIsActiveCondition+` AND
					groups_groups.idGroupParent = groups.ID AND groups_groups.idGroupChild = ?`, user.SelfGroupID).
		Where("groups_ancestors.ID IS NOT NULL OR groups_descendants.ID IS NOT NULL OR groups.bFreeAccess").
		Where("groups.ID = ?", groupID).Select(
		`groups.ID, groups.sName, groups.iGrade, groups.sDescription, groups.sDateCreated,
			groups.sType, groups.sRedirectPath, groups.bOpened, groups.bFreeAccess,
			IF(groups_ancestors.ID IS NOT NULL, groups.sCode, NULL) AS sCode,
			IF(groups_ancestors.ID IS NOT NULL, groups.sCodeTimer, NULL) AS sCodeTimer,
			IF(groups_ancestors.ID IS NOT NULL, groups.sCodeEnd, NULL) AS sCodeEnd,
			groups.bOpenContest,
			groups_ancestors.ID IS NOT NULL AS bCurrentUserIsOwner,
			groups_groups.ID IS NOT NULL AS bCurrentUserIsMember`).Limit(1)

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())

	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	render.Respond(w, r, service.ConvertMapFromDBToJSON(result[0]))

	return service.NoError
}
