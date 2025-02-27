package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /current-user/group-leave-requests/{group_id}/withdraw group-memberships groupLeaveRequestWithdraw
//
//	---
//	summary: Withdraw a group leave request
//	description: >
//		Lets the current user withdraw a request to leave a group (identified by {group_id}).
//
//
//		On success the service removes a row  with `group_id` = `{group_id}`,
//		`type` = 'leave_request' and `member_id` = user's `group_id` from the `group_pending_requests` table
//		and creates a new row in `group_membership_changes`
//		with `action` = `leave_request_withdrawn` and `at` equal to current UTC time.
//
//
//		The user should be a member of the group and there should be a row with
//		`type` = 'leave_request', `group_id` = `{parent_group_id}`
//		and `member_id` = user's `group_id` in `group_pending_requests`,
//		otherwise the "not found" error is returned.
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
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) withdrawGroupLeaveRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, withdrawGroupLeaveRequestAction)
}
