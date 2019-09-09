package currentuser

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-requests/{group_id} groups users groupRequestCreate
// ---
// summary: Create a request to join a group
// description: >
//   Lets a user create a request to join a group. There are two possible cases:
//
//   #### The user doesn't own the group
//
//     On success the service creates a new row in `groups_groups` with `idGroupParent` = user's self group ID,
//     `idGroupChild` = `group_id`, `groups_groups.sType` = `requestSent` and `sStatusDate` equal to current UTC time.
//
//     * `groups.bFreeAccess` should be 1, otherwise the 'forbidden' response is returned.
//
//     * If the group is a team with `idTeamItem` set and the user is already on a team with the same `idTeamItem`,
//       the unprocessable entity error is returned.
//
//     * If there is already a row in `groups_groups` with
//       `sType` = 'invitationSent'/'invitationAccepted'/'requestAccepted'/'joinedByCode'/'direct',
//       the unprocessable entity error is returned.
//
//     * If `groups_groups.sType` is `requestSent` already, the "unchanged" (201) response is returned.
//
//   #### The user owns the group
//
//     On success the service creates a new row in `groups_groups` with `idGroupParent` = user's self group ID,
//     `idGroupChild` = `group_id`, `groups_groups.sType` = `requestAccepted` and `sStatusDate` equal to current UTC time.
//
//     * If there is already a row in `groups_groups` with
//       `sType` = 'invitationAccepted'/'joinedByCode'/'direct',
//       the unprocessable entity error is returned.
//
//     * If `groups_groups.sType` is `requestAccepted` already, the "unchanged" (201) response is returned.
//
//     On success, the service propagates group ancestors in this case.
//
//
//   _Warning:_ The service doesn't check if the user has access rights on `idTeamItem` when the group is a team.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "201":
//     "$ref": "#/responses/createdOrUnchangedResponse"
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
