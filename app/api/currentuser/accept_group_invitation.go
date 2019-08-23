package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-invitations/{group_id}/accept groups users groupInvitationAccept
// ---
// summary: Accept a group invitation
// description:
//   Let a user approve an invitation to join a group.
//   On success the service sets `groups_groups.sType` to `invitationAccepted` and `sStatusDate` to current UTC time.
//   It also refreshes the access rights.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `ID` as a child with `sType`=`invitationSent`/`invitationAccepted`.
//     Otherwise the unprocessable entity error is returned.
//
//   * If `groups_groups.sType` is `invitationAccepted` already, the "unchanged" (200) response is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/updatedOrNotChangedResponse"
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
