package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model skillsViewResponseRow
type skillsViewResponseRow struct {
	*groupInfoForRootItem

	// required: true
	Skill *rootItem `json:"skill"`
}

// swagger:operation GET /current-user/group-memberships/skills group-memberships skillsView
// ---
// summary: List root skills
// description:
//   Returns the list of root skills of the groups the current user (or `{as_team_id}`) belongs to.
//
//
//   If `{as_team_id}` is given, it should be a user's parent team group, otherwise the "forbidden" error is returned.
// parameters:
// - name: as_team_id
//   in: query
//   type: integer
// responses:
//   "200":
//     description: OK. Success response with an array of root skills
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/skillsViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRootSkills(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getRootItems(w, r, false)
}
