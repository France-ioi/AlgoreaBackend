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
//   * If the group is a team with `team_item_id` set and the user is already on a team with the same `team_item_id`,
//     the unprocessable entity error is returned.
//
//   * There should be a row in `group_pending_requests` with the `{group_id}` as a parent as `group_id`
//     and the authenticated user’s `group_id` as `member_id` with `type`='invitation'.
//     Otherwise the unprocessable entity error is returned.
//
//   * If some of approvals required by the group are missing in `approvals`,
//     the unprocessable entity error is returned.
//
//
//   _Warning:_ The service doesn't check if the user has access rights on `team_item_id` when the group is a team.
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
//     "$ref": "#/responses/unprocessableEntityResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) acceptGroupInvitation(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, acceptInvitationAction)
}
