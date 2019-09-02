package groups

import (
	"net/http"

	"github.com/go-chi/render"

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
//   The authenticated user should be an owner of `group_id`, otherwise the 'forbidden' error is returned.
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
//       required: [id, name, grade, description, date_created, type, redirect_path, opened, free_access,
//                  code, code_timer, code_end, open_contest]
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

	query := srv.Store.Groups().OwnedBy(user).
		Where("groups.ID = ?", groupID).Select(
		`groups.ID, groups.sName, groups.iGrade, groups.sDescription, groups.sDateCreated,
     groups.sType, groups.sRedirectPath, groups.bOpened, groups.bFreeAccess,
     groups.sCode, groups.sCodeTimer, groups.sCodeEnd, groups.bOpenContest`).Limit(1)

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())

	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	render.Respond(w, r, service.ConvertMapFromDBToJSON(result[0]))

	return service.NoError
}
