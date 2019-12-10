package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/leave-requests/accept groups users groupLeaveRequestsAccept
// ---
// summary: Accept requests to leave a group
// description:
//   Lets an admin approve requests to leave a group.
//   On success the service removes `groups_groups` rows
//   with `parent_group_id` = `{parent_group_id}` and creates new `group_membership_changes` with
//   `group_id` = `{parent_group_id}`, `action` = 'leave_request_accepted`, `at` = current UTC time
//   for each of `group_ids`
//   The appropriate pending requests get removed from `group_pending_requests`.
//   The service also refreshes the access rights.
//
//
//   The authenticated user should be a manager of the `{parent_group_id}` with `can_manage` >= 'memberships',
//   otherwise the 'forbidden' error is returned.
//
//
//   There should be a row with `type` = 'leave_request' and `group_id` = `{parent_group_id}`
//   in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//   `invalid` as the result.
//
//
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
func (srv *Service) acceptLeaveRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performBulkMembershipAction(w, r, acceptLeaveRequestsAction)
}
