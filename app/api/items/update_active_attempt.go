package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) updateActiveAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	attemptID, err := service.ResolveURLQueryPathInt64Field(r, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	foundItemID, itemID, err := srv.Store.Attempts().GetAttemptItemIDIfUserHasAccess(attemptID, user)
	service.MustNotBeError(err)
	if !foundItemID {
		return service.InsufficientAccessRightsError
	}

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.UserItems().SetActiveAttempt(user.GroupID, itemID, attemptID))
		attemptStore := store.Attempts()
		service.MustNotBeError(
			attemptStore.ByID(attemptID).
				UpdateColumn(map[string]interface{}{
					"latest_activity_at": database.Now(),
				}).Error())
		service.MustNotBeError(attemptStore.ComputeAllAttempts())
		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
