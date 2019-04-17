package groups

import (
	"fmt"
	"github.com/go-chi/render"
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) removeChild(w http.ResponseWriter, r *http.Request) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	childGroupID, err := service.ResolveURLQueryPathInt64Field(r, "child_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	shouldDeleteOrphans := false
	if len(r.URL.Query()["delete_orphans"]) > 0 {
		shouldDeleteOrphans, err = service.ResolveURLQueryGetBoolField(r, "delete_orphans")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
	}

	user := srv.GetUser(r)
	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		apiErr = checkThatUserHasRightsForDirectRelation(s, user, parentGroupID, childGroupID)
		if apiErr != service.NoError {
			return apiErr.Error // rollback
		}

		// Check that the relation exists and it is a direct relation
		var result []struct{}
		service.MustNotBeError(s.GroupGroups().WithWriteLock().
			Where("idGroupParent = ?", parentGroupID).
			Where("idGroupChild = ?", childGroupID).
			Where("sType = 'direct'").Take(&result).Error())
		if len(result) == 0 {
			apiErr = service.InsufficientAccessRightsError
			return apiErr.Error // rollback
		}

		return s.GroupGroups().DeleteRelation(parentGroupID, childGroupID, shouldDeleteOrphans)
	})

	if apiErr != service.NoError {
		return apiErr
	}

	if err == database.ErrGroupBecomesOrphan {
		return service.ErrInvalidRequest(
			fmt.Errorf("group %d would become an orphan: confirm that you want to delete it", childGroupID))
	}

	service.MustNotBeError(err)
	service.MustNotBeError(render.Render(w, r, &service.Response{
		HTTPStatusCode: http.StatusOK,
		Success:        true,
		Message:        "deleted",
	}))
	return service.NoError
}
