package groups

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// for each login a result can be one of:
// "not_found" - login not found
// "invalid"   - when user is already a group member
//               (or there is a direct relation between the user
//                and the group... which should not happen)
// "unchanged" - if an invitation has been already sent
// "success"   - all other cases (note: user requests become accepted if any)

const maxAllowedLoginsToInvite = 100

// swagger:operation POST /groups/{parent_group_id}/invitations groups users groupInviteUsers
// ---
// summary: Invite users to a group
// description:
//   Lets an admin invite users, based on list of their logins, to join a group.
//   On success the service creates new rows in `groups_groups` with
//
//     * `sType` = "invitationSent"
//
//     * `sStatusDate` = current UTC time
//
//     * `idUserInviting` = `users.ID` of the authorized user,
//
//     * `sRole` = "member",
//
//     * correct `iChildOrder`, so that the row becomes the last child of the parent group.
//
//
//   It also refreshes the access rights.
//
//
//   * Logins not corresponding to valid users are ignored (result = "not_found").
//
//   * If the `parent_group_id` corresponds to a team with `idTeamItem` set, the service skips users
//     who are members of other teams with the same `idTeamItem` (result = "in_another_team").
//
//   * Pending group requests from users listed in `logins` become accepted (result = "success").
//
//   * Pending invitations stay unchanged (result = "unchanged).
//
//   * Group members (`groups_groups.sType` = "invitationAccepted"/"requestAccepted"/"direct")
//     are skipped (result = "invalid").
//
//
//   The action should not create cycles in the groups relations graph, otherwise
//   the login gets skipped with `cycle` as the result.
//
//
//   The response status code on success (201) doesn't depend on per-group results.
//
//
//   The authenticated user should be an owner of the `parent_group_id`, otherwise the 'forbidden' error is returned.
//
//
//   _Warning:_ The service doesn't check if the authenticated user or listed users have access rights
//   on `idTeamItem` when the `parent_group_id` represents a team.
// consumes:
// - application/json
// parameters:
// - name: parent_group_id
//   in: path
//   type: integer
//   required: true
// - in: body
//   name: logins_info
//   required: true
//   schema:
//     type: object
//     required: [logins]
//     properties:
//       logins:
//         type: array
//         items:
//           type: string
// responses:
//   "201":
//     "$ref": "#/responses/createdLoginRelationsResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) inviteUsers(w http.ResponseWriter, r *http.Request) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var requestData struct {
		Logins []string `json:"logins"`
	}
	err = render.Decode(r, &requestData)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	if len(requestData.Logins) == 0 {
		return service.ErrInvalidRequest(errors.New("there should be at least one login listed"))
	}
	if len(requestData.Logins) > maxAllowedLoginsToInvite {
		return service.ErrInvalidRequest(fmt.Errorf("there should be no more than %d logins", maxAllowedLoginsToInvite))
	}

	user := srv.GetUser(r)
	if apiErr := checkThatUserOwnsTheGroup(srv.Store, user, parentGroupID); apiErr != service.NoError {
		return apiErr
	}

	results := make(map[string]string, len(requestData.Logins))
	for _, login := range requestData.Logins {
		results[login] = notFound
	}

	var groupsToInviteRows []struct {
		Login       string `gorm:"column:sLogin"`
		SelfGroupID int64  `gorm:"column:idGroupSelf"`
	}
	service.MustNotBeError(srv.Store.Users().Select("sLogin, idGroupSelf").Where("sLogin IN (?)", requestData.Logins).
		Scan(&groupsToInviteRows).Error())

	groupsToInvite := make([]int64, 0, len(groupsToInviteRows))
	groupIDToLoginMap := make(map[int64]string, len(groupsToInviteRows))
	for _, row := range groupsToInviteRows {
		groupsToInvite = append(groupsToInvite, row.SelfGroupID)
		groupIDToLoginMap[row.SelfGroupID] = row.Login
	}

	var groupResults database.GroupGroupTransitionResults
	if len(groupsToInvite) > 0 {
		err = srv.Store.InTransaction(func(store *database.DataStore) error {
			groupsToInvite = filterOtherTeamsMembersOutForLogins(store, parentGroupID, groupsToInvite, results, groupIDToLoginMap)

			groupResults, err = store.GroupGroups().Transition(database.AdminCreatesInvitation, parentGroupID, groupsToInvite, user.ID)
			return err
		})
	}

	service.MustNotBeError(err)

	for id, result := range groupResults {
		results[groupIDToLoginMap[id]] = string(result)
	}

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(results)))

	return service.NoError
}

func filterOtherTeamsMembersOutForLogins(store *database.DataStore, parentGroupID int64, groupsToCheck []int64,
	results map[string]string, groupIDToLoginMap map[int64]string) []int64 {
	groupsToCheckMap := make(map[int64]bool, len(groupsToCheck))
	for _, id := range groupsToCheck {
		groupsToCheckMap[id] = true
	}

	otherTeamsMembers := getOtherTeamsMembers(store, parentGroupID, groupsToCheck)
	for _, id := range otherTeamsMembers {
		results[groupIDToLoginMap[id]] = inAnotherTeam
		delete(groupsToCheckMap, id)
	}
	newGroupsList := make([]int64, 0, len(groupsToCheckMap))
	for _, id := range groupsToCheck {
		if groupsToCheckMap[id] {
			newGroupsList = append(newGroupsList, id)
		}
	}
	return newGroupsList
}

func getOtherTeamsMembers(store *database.DataStore, parentGroupID int64, groupsToCheck []int64) []int64 {
	var parentGroupInfo struct {
		Type       string `gorm:"column:sType"`
		TeamItemID *int64 `gorm:"column:idTeamItem"`
	}
	const teamType = "Team"
	service.MustNotBeError(store.Groups().ByID(parentGroupID).WithWriteLock().Select("sType, idTeamItem").
		Take(&parentGroupInfo).Error())
	if parentGroupInfo.Type != teamType || parentGroupInfo.TeamItemID == nil {
		return nil
	}
	var otherTeamsMembers []int64
	service.MustNotBeError(store.Groups().TeamsMembersForItem(groupsToCheck, *parentGroupInfo.TeamItemID).WithWriteLock().
		Where("groups.ID != ?", parentGroupID).
		Pluck("idGroupChild", &otherTeamsMembers).Error())
	return otherTeamsMembers
}
