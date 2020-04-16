package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-invitations/{group_id}/accept group-memberships groupInvitationAccept
// ---
// summary: Accept a group invitation
// description:
//   Let a user approve an invitation to join a group.
//   On success the service creates a new `groups_groups` row
//   with `parent_group_id` = `group_id` and `child_group_id` = `user.group_id`,
//   and a new `group_membership_changes` row with `action` = 'invitation_accepted'
//   (the `at` field of both rows is set to current UTC time).
//   The invitation gets removed from `group_pending_requests`.
//   The service also refreshes the access rights.
//
//   * If the group is a team and the user is already on a team that has attempts for same contest
//     while the contest doesn't allow multiple attempts or that has active attempts for the same contest,
//     or if the group membership is frozen,
//     the unprocessable entity error is returned.
//
//   * There should be a row in `group_pending_requests` with the `{group_id}` as a parent as `group_id`
//     and the authenticated user’s `group_id` as `member_id` with `type`='invitation'.
//     Otherwise the "not found" error is returned.
//
//   * If some of approvals required by the group are missing in `approvals`,
//     the unprocessable entity error is returned with a list of missing approvals.
//
//   * If the group doesn't exist or is a user, the "forbidden" response is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: approvals
//   in: query
//   type: array
//   items:
//     type: string
//     enum: [personal_info_view,lock_membership,watch]
// responses:
//   "200":
//     "$ref": "#/responses/updatedOrUnchangedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponseWithMissingApprovals"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) acceptGroupInvitation(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, acceptInvitationAction)
}
