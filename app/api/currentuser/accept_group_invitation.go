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
//   On success the service sets `groups_groups.type` to `invitationAccepted` and `type_changed_at` to current UTC time.
//   It also refreshes the access rights.
//
//   * If the group is a team with `team_item_id` set and the user is already on a team with the same `team_item_id`,
//     the unprocessable entity error is returned.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `id` as a child with `type`=`invitationSent`/`invitationAccepted`.
//     Otherwise the unprocessable entity error is returned.
//
//   * If `groups_groups.type` is `invitationAccepted` already, the "unchanged" (200) response is returned.
//
//
//   _Warning:_ The service doesn't check if the user has access rights on `team_item_id` when the group is a team.
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
func (srv *Service) acceptGroupInvitation(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, acceptInvitationAction)
}
