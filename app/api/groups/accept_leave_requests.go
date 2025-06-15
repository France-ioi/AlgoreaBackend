package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/leave-requests/accept group-memberships groupLeaveRequestsAccept
//
//	---
//	summary: Accept group leave request
//	description:
//		Lets an admin approve user requests (identified by `{group_ids}`) to leave a group (identified by {parent_group_id}).
//		On success the service removes `groups_groups` rows
//		with `parent_group_id` = `{parent_group_id}` and creates new `group_membership_changes` with
//		`group_id` = `{parent_group_id}`, `action` = 'leave_request_accepted`, `at` = current UTC time
//		for each of `group_ids`
//		The appropriate pending requests get removed from `group_pending_requests`.
//		The service also refreshes the access rights.
//
//
//		The authenticated user should be a manager of the `{parent_group_id}` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned.  If the group is a user or the group membership is frozen,
//		the 'forbidden' error is returned as well.
//
//
//		There should be a row with `type` = 'leave_request' and `group_id` = `{parent_group_id}`
//		in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//		`invalid` as the result.
//
//
//		If the `{parent_group_id}` corresponds to a team, `{group_ids}` can contain no more than one id,
//		otherwise the 'bad request' response is returned.
//
//
//		If the `{parent_group_id}` corresponds to a team, the service skips a user with result = "in_another_team"
//		if removal breaks entry conditions of at least one of the team's participations
//		(i.e. any of `entry_min_admitted_members_ratio` or `entry_max_team_size` would not be satisfied).
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
func (srv *Service) acceptLeaveRequests(w http.ResponseWriter, r *http.Request) *service.APIError {
	return srv.performBulkMembershipAction(w, r, acceptLeaveRequestsAction)
}
