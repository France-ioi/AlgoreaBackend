package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/leave-requests/reject group-memberships groupLeaveRequestsReject
//
//	---
//	summary: Reject group leave requests
//	description:
//		Lets an admin reject requests (of users with ids in {group_ids}) to leave a group (identified by {parent_group_id}).
//		On success the service removes rows with `type` = 'leave_request' from `group_pending_requests` and
//		creates new rows with `action` = 'leave_request_refused' and `at` = current UTC time in `group_membership_changes`
//		for each of `group_ids`.
//
//
//		The authenticated user should be a manager of the `parent_group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//
//
//		There should be a row with `type` = 'leave_request' and `group_id` = `{parent_group_id}`
//		in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//		`invalid` as the result.
//
//
//		The response status code on success (200) doesn't depend on per-group results.
//	parameters:
//		- name: parent_group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: group_ids
//			in: query
//			type: array
//			items:
//				type: integer
//				format: int64
//			required: true
//	responses:
//		"200":
//			"$ref": "#/responses/updatedGroupRelationsResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) rejectLeaveRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performBulkMembershipAction(w, r, rejectLeaveRequestsAction)
}
