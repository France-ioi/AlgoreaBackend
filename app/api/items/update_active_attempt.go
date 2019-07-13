package items

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

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
			Where("idUser = ?", user.ID).Where("idItem = ?", itemID).
			UpdateColumn(map[string]interface{}{
				"idAttemptActive":            groupsAttemptID,
				"sLastActivityDate":          gorm.Expr("NOW()"),
				"sAncestorsComputationState": "todo",
			}).Error())
		service.MustNotBeError(store.GroupAttempts().
			ByID(groupsAttemptID).
			UpdateColumn(map[string]interface{}{
				"sLastActivityDate": gorm.Expr("NOW()"),
			}).Error())
		service.MustNotBeError(userItemStore.ComputeAllUserItems())
		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
