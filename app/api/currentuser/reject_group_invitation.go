package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-invitations/{group_id}/reject groups groupInvitationReject
// ---
// summary: Reject a group invitation
// description:
//   Let a user reject an invitation to join a group.
//   On success the service removes a `groups_pending_request` row
//   with `group_id` = `{group_id}` and `member_id` = `user.group_id`,
//   and adds a new `group_membership_changes` row with `action` = 'invitation_refused'
//   and `at` = current UTC time.
//
//   * There should be a row in `group_pending_requests` with the `{group_id}` as `group_id`
//   and the authenticated user’s `group_id` as `member_id` with `type`=`invitation_created`.
//   Otherwise the unprocessable entity error is returned.
//
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
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
func (srv *Service) rejectGroupInvitation(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, rejectInvitationAction)
}
