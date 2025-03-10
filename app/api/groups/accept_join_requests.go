package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /groups/{parent_group_id}/join-requests/accept group-memberships groupJoinRequestsAccept
//
//	---
//	summary: Accept group join requests
//	description:
//		Lets an admin approve user requests (identified by `{group_ids}`) to join a group (identified by {parent_group_id}).
//		On success the service creates new `groups_groups` rows
//		with `parent_group_id` = `{parent_group_id}` and new `group_membership_changes` with
//		`group_id` = `{parent_group_id}`, `action` = 'join_request_accepted`, `at` = current UTC time
//		for each of `group_ids`. The `groups_groups.*_approved_at` fields are set to `group_pending_requests.at`
//		for each approval given in the pending join requests.
//		Then the appropriate pending requests get removed from `group_pending_requests`.
//		The service also refreshes the access rights.
//
//
//		The authenticated user should be a manager of the `{parent_group_id}` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user or if the group membership is frozen,
//		the 'forbidden' error is returned as well.
//
//
//		If the `{parent_group_id}` corresponds to a team, `{group_ids}` can contain no more than one id,
//		otherwise the 'bad request' response is returned.
//
//
//		If the `{parent_group_id}` corresponds to a team, the service skips a user
//		being a member of another team having attempts for the same contest as `{parent_group_id}`
//		(expired attempts are ignored for contests allowing multiple attempts, result = "in_another_team").
//
//
//		If the `{parent_group_id}` corresponds to a team, the service skips a user with result = "in_another_team"
//		if joining breaks entry conditions of at least one of the team's participations
//		(i.e. any of `entry_min_admitted_members_ratio` or `entry_max_team_size` would not be satisfied).
//
//
//		There should be a row with `type` = 'join_request' and `group_id` = `{parent_group_id}`
//		in `group_pending_requests` for each of the input `group_ids`, otherwise the `group_id` gets skipped with
//		'invalid' as the result.
//
//
//		If the `{parent_group_id}` requires any approvals, but the pending request doesn't contain them,
//		the `group_id` gets skipped with 'approvals_missing' as the result.
//
//
//		The action should not create cycles in the groups relations graph, otherwise
//		the `group_id` gets skipped with `cycle` as the result.
//
//
//		If `groups.enforce_max_participants` is true and the new number of participants exceeds `groups.max_participants`
//		for the `{parent_group_id}` group,
//		all the valid joining groups get skipped with `full` as the result.
//		(The number of participants is computed as the number of non-expired users or teams which are direct children
//		of the group + invitations (join requests are not counted)).
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
func (srv *Service) acceptJoinRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performBulkMembershipAction(w, r, acceptJoinRequestsAction)
}
