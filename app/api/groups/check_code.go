package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupCodeCheckResponse
type groupCodeCheckResponse struct {
	// required:true
	Valid bool `json:"valid"`
}

// swagger:operation GET /groups/is-code-valid groups groupsCodeCheck
// ---
// summary: Check if the group code is valid
// description: >
//   Checks if it is possible for the current user (or for a new user if the current user is temporary)
//   to join a group with the given code.
//   The service returns false:
//
//   * if there is no team with `is_public` = 1, `code_expires_at` > NOW() (or NULL), and `code` = `{code}`;
//
//   * if the group is a team and the user is already on a team that has attempts for same contest
//     while the contest doesn't allow multiple attempts or that has active attempts for the same contest,
//     or if the group membership is frozen;
//
//   * if there is already a row in `groups_groups` with the found team as a parent
//     and the user’s selfGroup’s id as a child;
//
//   * if the group is a team and joining breaks entry conditions of at least one of the team's participations
//     (i.e. any of `entry_min_admitted_members_ratio` or `entry_max_team_size` would not be satisfied).
//
//   Otherwise, the service returns true.
// parameters:
// - name: code
//   in: query
//   type: string
//   required: true
// responses:
//   "200":
//     description: OK. Validity of the code
//     schema:
//       "$ref": "#/definitions/groupCodeCheckResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) checkCode(w http.ResponseWriter, r *http.Request) service.APIError {
	code, err := service.ResolveURLQueryGetStringField(r, "code")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	userIDToCheck := user.GroupID
	if user.IsTempUser {
		userIDToCheck = domain.ConfigFromContext(r.Context()).AllUsersGroupID
	}

	response := groupCodeCheckResponse{
		Valid: checkGroupCodeForUser(srv.Store, userIDToCheck, code),
	}

	render.Respond(w, r, &response)
	return service.NoError
}

func checkGroupCodeForUser(store *database.DataStore, userIDToCheck int64, code string) bool {
	info, err := store.GetTeamJoiningByCodeInfoByCode(code, false)
	service.MustNotBeError(err)
	if info == nil || info.FrozenMembership {
		return false
	}

	found, err := store.CheckIfTeamParticipationsConflictWithExistingUserMemberships(info.TeamID, userIDToCheck, false)
	service.MustNotBeError(err)
	if found {
		return false
	}

	ok, err := store.Groups().CheckIfEntryConditionsStillSatisfiedForAllActiveParticipations(info.TeamID, userIDToCheck, true, false)
	service.MustNotBeError(err)
	return ok
}
