package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user/group-requests/{group_id} group-memberships groupJoinRequestWithdraw
// ---
// summary: Withdraw a request to join a group
// description: >
//   Lets a user withdraw a request to join a group.
//
//
//     On success the service removes a row  with
//     `group_id` = `{group_id}`, `member_id` = user's self group id and `type` = 'join_request'
//     from the `group_pending_requests` table,
//     and creates a new row in `group_membership_changes` for the same group-user pair
//     with `action` = 'join_request_withdrawn' and `at` equal to current UTC time.
//
//     * If there is no row in `group_pending_requests` for the group-user pair with
//       `type` == 'join_request', the "not found" error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "201":
//     "$ref": "#/responses/createdOrUnchangedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) withdrawGroupJoinRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, withdrawGroupJoinRequestAction)
}
