package groups

import (
	"errors"
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

	user := srv.GetUser(r)
	if apiErr := checkThatUserOwnsTheGroup(srv.Store, user, parentGroupID); apiErr != service.NoError {
		return apiErr
	}

	results := make(map[string]string, len(requestData.Logins))
	for _, login := range requestData.Logins {
		results[login] = "not_found"
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
			groupResults, err = store.GroupGroups().Transition(database.AdminCreatesInvitation, parentGroupID, groupsToInvite, user.UserID)
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
