package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /current-user/group-requests/{group_id}/withdraw group-memberships groupJoinRequestWithdraw
//
//	---
//	summary: Withdraw a group join request
//	description: >
//		Lets the current user withdraw a request to join a group (idenfified by {group_id}).
//
//
//		On success the service removes a row  with
//		`group_id` = `{group_id}`, `member_id` = user's self group id and `type` = 'join_request'
//		from the `group_pending_requests` table,
//		and creates a new row in `group_membership_changes` for the same group-user pair
//		with `action` = 'join_request_withdrawn' and `at` equal to current UTC time.
//
//		* If there is no row in `group_pending_requests` for the group-user pair with
//			`type` == 'join_request', the "not found" error is returned.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"422":
//			"$ref": "#/responses/unprocessableEntityResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) withdrawGroupJoinRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, withdrawGroupJoinRequestAction)
}
