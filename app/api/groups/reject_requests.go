package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/requests/reject groups users groupRequestsReject
// ---
// summary: Reject requests to join a group
// description:
//   Lets an admin reject requests to join a group.
//   On success the service sets `groups_groups.sType` to "requestRefused" and `sStatusDate` to current UTC time
//   for each of `group_ids`.
//
//
//   The authenticated user should be an owner of the `parent_group_id`, otherwise the 'forbidden' error is returned.
//
//
//   The input `group_id` should have the input `parent_group_id` as a parent group and the
//   `groups_groups.sType` should be "requestSent", otherwise the `group_id` gets skipped with
//   `unchanged` (if `sType` = "requestRefused") or `invalid` as the result.
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
func (srv *Service) rejectRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.acceptOrRejectRequests(w, r, rejectRequestsAction)
}
