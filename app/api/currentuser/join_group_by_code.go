package currentuser

import (
	"errors"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /current-user/group-memberships/by-code group-memberships groupsJoinByCode
// ---
// summary: Join a team using a code
// description:
//   Lets a user to join a team group by a code.
//   On success the service inserts a row into `groups_groups`
//   with `parent_group_id` = `id` of the team found by the code and `child_group_id` = `group_id` of the user
//   and another row into `group_membership_changes`
//   with `group_id` = `id` of the team, `member_id` = `group_id` of the user, `action`=`joined_by_code`,
//   and `at` = current UTC time.
//   It also refreshes the access rights.
//
//   * If there is no team with `is_public` = 1, `code_expires_at` > NOW() (or NULL), and `code` = `code`,
//     or if the current user is temporary, the forbidden error is returned.
//
//   * If the group is a team and the user is already on a team that has attempts for same contest
//     while the contest doesn't allow multiple attempts or that has active attempts for the same contest,
//     or if the group membership is frozen,
//     the unprocessable entity error is returned.
//
//   * If there is already a row in `groups_groups` with the found team as a parent
//     and the authenticated user’s selfGroup’s id as a child, the unprocessable entity error is returned.
//
//   * If the group is a team and joining breaks entry conditions of at least one of the team's participations
//     (i.e. any of `entry_min_admitted_members_ratio` or `entry_max_team_size` would not be satisfied),
//     the unprocessable entity error is returned.
//
//   * If the group requires some approvals from the user and those are not given in `approval`,
//     the unprocessable entity error is returned with a list of missing approvals.
// parameters:
// - name: code
//   in: query
//   type: string
//   required: true
// - name: approvals
//   in: query
//   type: array
//   items:
//     type: string
//     enum: [personal_info_view,lock_membership,watch]
// responses:
//   "201":
//     description: Created. The request has successfully created the group relation.
//     schema:
//       "$ref": "#/definitions/createdResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "422":
//     "$ref": "#/responses/unprocessableEntityResponseWithMissingApprovals"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) joinGroupByCode(w http.ResponseWriter, r *http.Request) service.APIError {
	code, err := service.ResolveURLQueryGetStringField(r, "code")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	if user.IsTempUser {
		return service.InsufficientAccessRightsError
	}

	apiError := service.NoError
	var results database.GroupGroupTransitionResults
	var approvalsToRequest map[int64]database.GroupApprovals
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		info, errInTransaction := store.GetTeamJoiningByCodeInfoByCode(code, true)
		service.MustNotBeError(errInTransaction)
		if info == nil {
			logging.GetLogEntry(r).Warnf("A user with group_id = %d tried to join a group using a wrong/expired code", user.GroupID)
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}

		if info.FrozenMembership {
			apiError = service.ErrUnprocessableEntity(errors.New("group membership is frozen"))
			return apiError.Error // rollback
		}

		var found bool
		found, errInTransaction = store.CheckIfTeamParticipationsConflictWithExistingUserMemberships(info.TeamID, user, true)
		service.MustNotBeError(errInTransaction)
		if found {
			apiError = service.ErrUnprocessableEntity(errors.New("team's participations are in conflict with the user's participations"))
			return apiError.Error // rollback
		}

		var ok bool
		ok, errInTransaction = store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(info.TeamID, user.GroupID, true)
		service.MustNotBeError(errInTransaction)
		if !ok {
			apiError = service.ErrUnprocessableEntity(errors.New("entry conditions would not be satisfied"))
			return apiError.Error // rollback
		}

		if info.CodeExpiresAtIsNull && !info.CodeLifetimeIsNull {
			service.MustNotBeError(store.Groups().ByID(info.TeamID).
				UpdateColumn("code_expires_at", gorm.Expr("ADDTIME(NOW(), code_lifetime)")).Error())
		}

		var approvals database.GroupApprovals
		approvals.FromString(r.URL.Query().Get("approvals"))
		results, approvalsToRequest, errInTransaction = store.GroupGroups().Transition(
			database.UserJoinsGroupByCode, info.TeamID, []int64{user.GroupID},
			map[int64]database.GroupApprovals{user.GroupID: approvals}, user.GroupID)
		return errInTransaction
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	return RenderGroupGroupTransitionResult(w, r, results[user.GroupID], approvalsToRequest[user.GroupID], joinGroupByCodeAction)
}
