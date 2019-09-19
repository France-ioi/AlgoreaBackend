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
		userItemStore := store.UserItems()
		service.MustNotBeError(userItemStore.CreateIfMissing(user.ID, itemID))
		service.MustNotBeError(userItemStore.
			Where("user_id = ?", user.ID).Where("item_id = ?", itemID).
			UpdateColumn(map[string]interface{}{
				"attempt_active_id":           groupsAttemptID,
				"last_activity_date":          database.Now(),
				"ancestors_computation_state": "todo",
			}).Error())
		service.MustNotBeError(store.GroupAttempts().
			ByID(groupsAttemptID).
			UpdateColumn(map[string]interface{}{
				"last_activity_date": database.Now(),
			}).Error())
		service.MustNotBeError(userItemStore.ComputeAllUserItems())
		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
