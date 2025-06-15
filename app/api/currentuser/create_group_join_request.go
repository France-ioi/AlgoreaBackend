package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /current-user/group-requests/{group_id} group-memberships groupJoinRequestCreate
//
//	---
//	summary: Create a group join request
//	description: >
//		Lets the current user create a request to join a group (idenfified by {group_id}).
//		There are two possible cases:
//
//		#### The user is not a manager of the group
//
//		On success the service creates a new row in `group_pending_requests` with
//		`group_id` = `{group_id}`, `member_id` = user's self group id, `type` = 'join_request',
//		given `approvals` and `at` equal to current UTC time,
//		and a new row in `group_membership_changes` for the same pair of groups
//		with `action` = 'join_request_created' and `at` equal to current UTC time.
//
//		* `groups.is_public` should be 1, otherwise the 'forbidden' response is returned.
//
//		* If there is already a row in `group_pending_requests` with
//			`type` != 'join_request' or a row in `groups_groups` for the same group-user pair,
//			the unprocessable entity error is returned.
//
//		* If there is already a row in `group_pending_requests` with `type` = 'join_request',
//			the "unchanged" (201) response is returned.
//
//		#### The user is a manager of the group with `can_manage` >= 'memberships'
//
//		On success the service creates a new row in `groups_groups` with `parent_group_id` = `group_id`,
//		given `approvals` and `child_group_id` = user's self group id + a new row in `group_membership_changes`
//		for the same group pair with `action` = `join_request_accepted` and `at` equal to current UTC time.
//		A pending request/invitation gets removed from `group_pending_requests`.
//
//		* If there is already a row in `groups_groups` or a row in `group_pending_request` with
//			`type` != 'invitation'/'join_request', the unprocessable entity error is returned.
//
//		On success, the service propagates group ancestors in this case.
//
//
//		In both cases, if some approvals required by the group are missing in `approvals`,
//		the unprocessable entity error with a list of missing approvals is returned.
//
//
//		If the group doesn't exist, or it is a user, or its membership is frozen, or the current user is a temporary user,
//		the "forbidden" response is returned.
//
//
//		If the group is a team and the user is already on a team that has attempts for the same item requiring explicit entry
//		while the item doesn't allow multiple attempts or that has active attempts for the same item requiring explicit entry,
//		the unprocessable entity error is returned.
//
//
//		If the group is a team and joining breaks entry conditions of at least one of the team's participations
//		(i.e. any of `entry_min_admitted_members_ratio` or `entry_max_team_size` would not be satisfied),
//		the unprocessable entity error is returned.
//
//
//		If `groups.enforce_max_participants` is true and the number of participants >= `groups.max_participants`,
//		the conflict error is returned.
//		(The number of participants is computed as the number of non-expired users or teams which are direct children
//		 of the group + invitations (join requests are not counted)).
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: approvals
//			in: query
//			type: array
//			items:
//				type: string
//				enum: [personal_info_view,lock_membership,watch]
//	responses:
//		"201":
//			"$ref": "#/responses/createdOrUnchangedResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"409":
//			"$ref": "#/responses/conflictResponse"
//		"422":
//			"$ref": "#/responses/unprocessableEntityResponseWithMissingApprovals"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroupJoinRequest(w http.ResponseWriter, r *http.Request) *service.APIError {
	return srv.performGroupRelationAction(w, r, createGroupJoinRequestAction)
}
