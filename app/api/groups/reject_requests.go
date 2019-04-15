package groups

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) rejectRequests(w http.ResponseWriter, r *http.Request) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "group_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	if apiErr := checkThatUserOwnsTheGroup(srv.Store, user, parentGroupID); apiErr != service.NoError {
		return apiErr
	}

	var results database.GroupGroupTransitionResults
	if len(groupIDs) > 0 {
		err = srv.Store.InTransaction(func(store *database.DataStore) error {
			results, err = store.GroupGroups().Transition(database.AdminRefusesRequest, parentGroupID, groupIDs)
			return err
		})
	}

	service.MustNotBeError(err)

	renderGroupGroupTransitionResults(w, r, results)
	return service.NoError
}
