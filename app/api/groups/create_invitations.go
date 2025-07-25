package groups

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// for each login a result can be one of:
// "not_found" - login not found
// "invalid"   - when user is already a group member
//               (or there is a direct relation between the user
//                and the group... which should not happen)
// "unchanged" - if an invitation has been already sent
// "full"      - for all the invitations if `groups.enforce_max_participants` is true and
//               the new total number of non-expired users or teams
//               which are direct children of the group + invitations > `groups.max_participants`
// "success"   - all other cases (note: user requests become accepted if any)

const maxAllowedLoginsToInvite = 100

// swagger:operation POST /groups/{parent_group_id}/invitations group-memberships groupInvitationsCreate
//
//	---
//	summary: Create group invitations
//	description:
//		Lets an admin invite users, based on list of their logins, to join a group (identified by {parent_group_id}).
//		On success the service creates new rows in `group_pending_requests` with
//
//			* `type` = "invitation"
//
//			* `at` = current UTC time
//
//		and `group_membership_changes` with
//
//			* `action` = "invitation_created"
//
//			* `at` = current UTC time.
//
//			* `initiator_id` = `users.group_id` of the authorized user.
//
//
//		It also refreshes the access rights when needed.
//
//
//		* Logins not corresponding to valid users or corresponding to temporary users are ignored (result = "not_found").
//
//		* If the `parent_group_id` corresponds to a team, the service skips users
//			who are members of other teams participating in solving same items requiring explicit entry as `parent_group_id`
//			(expired/ended attempts are ignored for items allowing multiple attempts, result = "in_another_team").
//
//		* Pending group requests from users listed in `logins` become accepted (result = "success")
//			if all needed approvals are given, or replaced by invitations otherwise.
//
//		* Pending invitations stay unchanged (result = "unchanged).
//
//		* Group members (already having `groups_groups`) are skipped (result = "invalid").
//
//
//		The action should not create cycles in the groups relations graph, otherwise
//		the login gets skipped with `cycle` as the result.
//
//
//		If `groups.enforce_max_participants` is true and the new number of participants exceeds `groups.max_participants`,
//		all the valid logins get skipped with `full` as the result.
//		(The number of participants is computed as the number of non-expired users or teams which are direct children
//		of the group + invitations (join requests are not counted)).
//
//
//		The response status code on success (201) doesn't depend on per-group results.
//
//
//		The authenticated user should be a manager of the `parent_group_id` with `can_manage` >= 'memberships',
//		otherwise the 'forbidden' error is returned. If the group is a user, the 'forbidden' error is returned as well.
//	consumes:
//		- application/json
//	parameters:
//		- name: parent_group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- in: body
//			name: logins_info
//			required: true
//			schema:
//				type: object
//				required: [logins]
//				properties:
//					logins:
//						type: array
//						items:
//							type: string
//	responses:
//		"201":
//			"$ref": "#/responses/createdLoginRelationsResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) createGroupInvitations(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var requestData struct {
		Logins []string `json:"logins"`
	}
	err = render.Decode(httpRequest, &requestData)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	if len(requestData.Logins) == 0 {
		return service.ErrInvalidRequest(errors.New("there should be at least one login listed"))
	}
	if len(requestData.Logins) > maxAllowedLoginsToInvite {
		return service.ErrInvalidRequest(fmt.Errorf("there should be no more than %d logins", maxAllowedLoginsToInvite))
	}

	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)
	service.MustNotBeError(checkThatUserCanManageTheGroupMemberships(store, user, parentGroupID))

	results := make(map[string]string, len(requestData.Logins))
	for _, login := range requestData.Logins {
		results[login] = notFound
	}

	var groupsToInviteRows []struct {
		Login   string
		GroupID int64
	}
	service.MustNotBeError(store.Users().Select("login, group_id").
		Where("login IN (?)", requestData.Logins).
		Where("NOT temp_user").
		Scan(&groupsToInviteRows).Error())

	groupsToInvite := make([]int64, 0, len(groupsToInviteRows))
	groupIDToLoginMap := make(map[int64]string, len(groupsToInviteRows))
	for _, row := range groupsToInviteRows {
		groupsToInvite = append(groupsToInvite, row.GroupID)
		groupIDToLoginMap[row.GroupID] = row.Login
	}

	var groupResults database.GroupGroupTransitionResults
	if len(groupsToInvite) > 0 {
		err = store.InTransaction(func(store *database.DataStore) error {
			groupsToInvite = filterOtherTeamsMembersOutForLogins(store, parentGroupID, groupsToInvite, results, groupIDToLoginMap)

			groupResults, _, err = store.GroupGroups().
				Transition(database.AdminCreatesInvitation, parentGroupID, groupsToInvite, nil, user.GroupID)
			return err
		})
	}

	service.MustNotBeError(err)

	for id, result := range groupResults {
		results[groupIDToLoginMap[id]] = string(result)
	}

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.CreationSuccess(results)))

	return nil
}

func filterOtherTeamsMembersOutForLogins(store *database.DataStore, parentGroupID int64, groupsToCheck []int64,
	results map[string]string, groupIDToLoginMap map[int64]string,
) []int64 {
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
	found, err := store.Groups().ByID(parentGroupID).Where("type = 'Team'").WithExclusiveWriteLock().HasRows()
	service.MustNotBeError(err)
	if !found {
		return nil
	}

	teamAttemptsQuery := store.Attempts().
		Where("participant_id = ?", parentGroupID).
		Where("root_item_id IS NOT NULL").
		Group("root_item_id").WithExclusiveWriteLock()

	var otherTeamsMembers []int64
	service.MustNotBeError(store.ActiveGroupGroups().Where("child_group_id IN (?)", groupsToCheck).
		Joins("JOIN (?) AS team_attempts", // all the team's attempts (not only active ones)
			teamAttemptsQuery.
				Select(`
					root_item_id AS item_id,
					MAX(NOW() < attempts.allows_submissions_until AND attempts.ended_at IS NULL) AS is_active`).QueryExpr()).
		Joins("JOIN items ON items.id = team_attempts.item_id").
		Joins("JOIN attempts ON attempts.participant_id = groups_groups_active.parent_group_id AND attempts.root_item_id = items.id").
		Where("groups_groups_active.is_team_membership = 1").
		Where("parent_group_id != ?", parentGroupID).
		Where(`
			(team_attempts.is_active AND NOW() < attempts.allows_submissions_until AND attempts.ended_at IS NULL) OR
			NOT items.allows_multiple_attempts`).
		WithExclusiveWriteLock().Pluck("child_group_id", &otherTeamsMembers).Error())

	return otherTeamsMembers
}
