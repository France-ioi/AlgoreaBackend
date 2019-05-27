package items

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) updateActiveAttempt(w http.ResponseWriter, r *http.Request) service.APIError {
	groupsAttemptID, err := service.ResolveURLQueryGetInt64Field(r, "groups_attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	selfGroupID, err := user.SelfGroupID()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var itemID int64
	usersGroupsQuery := srv.Store.GroupGroups().WhereUserIsMember(user).Select("idGroupParent")
	err = srv.Store.Items().Visible(user).
		Joins("JOIN groups_attempts ON groups_attempts.idItem = items.ID AND groups_attempts.ID = ?", groupsAttemptID).
		Joins("JOIN users_items ON users_items.idItem = items.ID AND users_items.idUser = ?", user.UserID).
		Where("partialAccess > 0 OR fullAccess > 0").
		Where("IF(items.bHasAttempts, groups_attempts.idGroup IN ?, groups_attempts.idGroup = ?)", usersGroupsQuery.SubQuery(), selfGroupID).
		PluckFirst("items.ID", &itemID).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(srv.Store.InTransaction(func(store *database.DataStore) error {
		service.MustNotBeError(store.UserItems().
			Where("idUser = ?", user.UserID).Where("idItem = ?", itemID).
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
		service.MustNotBeError(store.UserItems().ComputeAllUserItems())
		return nil
	}))

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}
