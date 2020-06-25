package service

import (
	"errors"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// GetParticipantIDFromRequest returns `as_team_id` parameter value if it is given or the user's `group_id` otherwise.
// If `as_team_id` is given, it should be an id of a team and the user should be a member of this team, otherwise
// the 'forbidden' error is returned.
func GetParticipantIDFromRequest(httpReq *http.Request, user *database.User, store *database.DataStore) (int64, APIError) {
	groupID := user.GroupID
	var err error
	if len(httpReq.URL.Query()["as_team_id"]) != 0 {
		groupID, err = ResolveURLQueryGetInt64Field(httpReq, "as_team_id")
		if err != nil {
			return 0, ErrInvalidRequest(err)
		}

		var found bool
		found, err = store.Groups().ByID(groupID).Where("type = 'Team'").
			Joins("JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id").
			Where("groups_groups_active.child_group_id = ?", user.GroupID).HasRows()
		MustNotBeError(err)
		if !found {
			return 0, ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}
	return groupID, NoError
}
