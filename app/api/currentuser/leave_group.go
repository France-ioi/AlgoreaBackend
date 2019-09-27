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
//   On success the service sets `groups_groups.type` to `left` and `status_changed_at` to current UTC time.
//   It also refreshes the access rights.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `id` as a child with `type`=`invitationAccepted`/`requestAccepted`/`direct`/`left`.
//     Otherwise the unprocessable entity error is returned.
//
//   * If `groups_groups.type` is `left` already, the "unchanged" (200) response is returned.
//
//   * The user cannot leave the group if `NOW()` < `groups.lock_user_deletion_until`.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/deletedOrUnchangedResponse"
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
