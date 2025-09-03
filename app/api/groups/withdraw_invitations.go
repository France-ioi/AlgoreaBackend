package groups

import (
	"net/http"
)

// swagger:operation POST /groups/{parent_group_id}/invitations/withdraw group-memberships groupInvitationsWithdraw
//
//	---
//	summary: Withdraw group invitations
//	description:
//		Lets a manager withdraw invitations (of users with ids in {group_ids}) to join a group (idenfified by {parent_group_id}).
//		On success the service removes rows with `type` = 'invitation' from `group_pending_requests` and
//		creates new rows with `action` = 'invitation_withdrawn' and `at` = current UTC time in `group_membership_changes`
//		for each of `group_ids`.
//
//
//		The authenticated user should be a manager of the `parent_group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//
//
//		There should be a row with `type` = 'invitation' and `group_id` = `{parent_group_id}`
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
func (srv *Service) withdrawInvitations(w http.ResponseWriter, r *http.Request) error {
	return srv.performBulkMembershipAction(w, r, withdrawInvitationsAction)
}
