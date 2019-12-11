package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation DELETE /current-user/group-memberships/{group_id} group-memberships groupLeave
// ---
// summary: Leave a group
// description:
//   Lets a user to leave a group.
//   On success the service removes a row with with `parent_group_id` = `group_id` and `child_group_id` = `user.group_id`
//   from `groups_groups`, and adds a new `group_membership_changes` row with `action` = 'left'
//   and `at` = current UTC time.
//   It also refreshes the access rights.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `id` as a child.
//     Otherwise the "not found" error is returned.
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
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) leaveGroup(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, leaveGroupAction)
}
