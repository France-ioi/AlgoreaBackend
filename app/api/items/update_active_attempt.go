package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) updateActiveAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	groupsAttemptID, err := service.ResolveURLQueryPathInt64Field(r, "groups_attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	foundItemID, itemID, err := srv.Store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(groupsAttemptID, user)
	service.MustNotBeError(err)
	if !foundItemID {
		return service.InsufficientAccessRightsError
	}

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.UserItems().SetActiveAttempt(user.ID, itemID, groupsAttemptID))
		groupAttemptStore := store.GroupAttempts()
		service.MustNotBeError(
			groupAttemptStore.ByID(groupsAttemptID).
				UpdateColumn(map[string]interface{}{
					"latest_activity_at": database.Now(),
				}).Error())
		service.MustNotBeError(groupAttemptStore.ComputeAllGroupAttempts())
		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
