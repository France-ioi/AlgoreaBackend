package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/requests/accept groups users groupRequestsAccept
// ---
// summary: Accept requests to join a group
// description:
//   Lets an admin approve requests to join a group.
//   On success the service sets `groups_groups.sType` to "requestAccepted" and `sStatusDate` to current UTC time
//   for each of `group_ids`.
//   It also refreshes the access rights.
//
//
//   The authenticated user should be an owner of the `parent_group_id`, otherwise the 'forbidden' error is returned.
//
//
//   If the `parent_group_id` corresponds to a team with `idTeamItem` set, the service skips users
//   who are members of other teams with the same `idTeamItem` (result = "in_another_team").
//
//
//   The input `group_id` should have the input `parent_group_id` as a parent group and the
//   `groups_groups.sType` should be "requestSent", otherwise the `group_id` gets skipped with
//   `unchanged` (if `sType` = "requestAccepted") or `invalid` as the result.
//
//
//   The action should not create cycles in the groups relations graph, otherwise
//   the `group_id` gets skipped with `cycle` as the result.
//   The response status code on success (200) doesn't depend on per-group results.
// parameters:
// - name: parent_group_id
//   in: path
//   type: integer
//   required: true
// - name: group_ids
//   in: query
//   type: array
//   items:
//     type: integer
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/updatedGroupRelationsResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) acceptRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.acceptOrRejectRequests(w, r, acceptRequestsAction)
}
