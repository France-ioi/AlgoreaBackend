package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-requests/{group_id} groups users groupRequestCreate
// ---
// summary: Send a request to join a group
// description:
//   Lets a user send a request to join a group.
//   On success the service creates a new row in `groups_groups` with `idGroupParent` = user's self group ID,
//   `idGroupChild` = `group_id`, `groups_groups.sType` = `requestSent` and `sStatusDate` equal to current UTC time.
//
//   * `groups.bFreeAccess` should be 1, otherwise the 'forbidden' response is returned.
//
//   * There should be a row in `groups_groups` with the `group_id` as a parent
//     and the authenticated user’s selfGroup’s `ID` as a child with `sType`=`invitationSent`/`invitationRefused`.
//     Otherwise the unprocessable entity error is returned.
//
//   * If `groups_groups.sType` = 'invitationSent'/'invitationAccepted'/'requestAccepted'/'direct',
//     the unprocessable entity error is returned
//
//   * If `groups_groups.sType` is `requestSent` already, the "not changed" (201) response is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "201":
//     "$ref": "#/responses/createdOrNotChangedResponse"
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
func (srv *Service) sendGroupRequest(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.performGroupRelationAction(w, r, createGroupRequestAction)
}
