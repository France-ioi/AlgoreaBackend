package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-leave-requests/{group_id} group-memberships groupLeaveRequestCreate
//
//		---
//		summary: Create a group leave request
//		description: >
//	    Lets the current user create a request to leave a group (idenfified by {group_id}).
//
//
//	    On success the service creates a new row in `group_pending_requests` with `group_id` = `{group_id}`
//	    `type` = 'leave_request' and `member_id` = user's `group_id` + a new row in `group_membership_changes`
//	    with `action` = `leave_request_created` and `at` equal to current UTC time.
//
//
//	    If there is already a row in `groups_groups` and a row in `group_pending_request` with
//	    `type` == 'leave_request', the "unchanged" (201) response is returned.
//
//
//	    The user should be a member of the `{group_id}` and
//	    the group's `require_lock_membership_approval_until` should be greater than NOW(),
//	    and `groups_groups.lock_membership_approved` should be set, and the group membership should not be frozen.
//	    Otherwise the "forbidden" error is returned.
//		parameters:
//			- name: group_id
//				in: path
//				type: integer
//				required: true
//		responses:
//			"201":
//				"$ref": "#/responses/createdOrUnchangedResponse"
//			"400":
//				"$ref": "#/responses/badRequestResponse"
//			"401":
//				"$ref": "#/responses/unauthorizedResponse"
//			"403":
//				"$ref": "#/responses/forbiddenResponse"
//			"422":
//				"$ref": "#/responses/unprocessableEntityResponse"
//			"500":
//				"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroupLeaveRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, createGroupLeaveRequestAction)
}
