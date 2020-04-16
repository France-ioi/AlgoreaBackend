package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/join-requests/accept group-memberships groupJoinRequestsAccept
// ---
// summary: Accept requests to join a group
// description:
//   Lets an admin approve requests to join a group.
//   On success the service creates new `groups_groups` rows
//   with `parent_group_id` = `{parent_group_id}` and new `group_membership_changes` with
//   `group_id` = `{parent_group_id}`, `action` = 'join_request_accepted`, `at` = current UTC time
//   for each of `group_ids`. The `groups_groups.*_approved_at` fields are set to `group_pending_requests.at`
//   for each approval given in the pending join requests.
//   Then the appropriate pending requests get removed from `group_pending_requests`.
//   The service also refreshes the access rights.
//
//
//   The authenticated user should be a manager of the `{parent_group_id}` with `can_manage` >= 'memberships',
//   otherwise the 'forbidden' error is returned. If the group is a user or if the group membership is frozen,
//   the 'forbidden' error is returned as well.
//
//
//   If the `{parent_group_id}` corresponds to a team, the service skips users
//   who are members of other teams having attempts for the same contests as `{parent_group_id}`
//   (expired attempts are ignored for contests allowing multiple attempts, result = "in_another_team").
//
//
//   There should be a row with `type` = 'join_request' and `group_id` = `{parent_group_id}`
//   in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//   'invalid' as the result.
//
//
//   If the `{parent_group_id}` requires any approvals, but the pending request doesn't contain them,
//   the `group_id` gets skipped with 'approvals_missing' as the result.
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
func (srv *Service) acceptJoinRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performBulkMembershipAction(w, r, acceptJoinRequestsAction)
}
