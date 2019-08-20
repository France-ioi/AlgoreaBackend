package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user/group-memberships/{group_id} groups users groupLeave
// ---
// summary: Leave a group
// description:
//   Lets a user to leave a group.
//   On success the service sets `groups_groups.sType` to `left` and `sStatusDate` to current UTC time.
//   It also refreshes the access rights.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `ID` as a child with `sType`=`invitationAccepted`/`requestAccepted`/`direct`/`left`.
//     Otherwise the unprocessable entity error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/deletedOrNotChangedResponse"
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
func (srv *Service) leaveGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, leaveGroupAction)
}
