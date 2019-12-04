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
//   On success the service creates new `groups_groups` rows
//   with `parent_group_id` = `{parent_group_id}` and new `group_membership_changes` with
//   `group_id` = `{parent_group_id}`, `action` = 'join_request_accepted`, `at` = current UTC time
//   for each of `group_ids`
//   The appropriate pending requests get removed from `group_pending_requests`.
//   The service also refreshes the access rights.
//
//
//   The authenticated user should be a manager of the `{parent_group_id}` with `can_manage` >= 'memberships',
//   otherwise the 'forbidden' error is returned.
//
//
//   If the `{parent_group_id}` corresponds to a team with `team_item_id` set, the service skips users
//   who are members of other teams with the same `team_item_id` (result = "in_another_team").
//
//
//   There should be a row with `type` = 'join_request' and `group_id` = `{parent_group_id}`
//   in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//   `invalid` as the result.
//
//
//   The action should not create cycles in the groups relations graph, otherwise
//   the `group_id` gets skipped with `cycle` as the result.
//   The response status code on success (200) doesn't depend on per-group results.
//
//
//   _Warning:_ The service doesn't check if the authenticated user or requesting users have access rights
//   on `team_item_id` when the `parent_group_id` represents a team.
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
