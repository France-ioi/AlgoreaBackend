package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) addChild(w http.ResponseWriter, r *http.Request) service.APIError {
	parentGroupID, err := service.ResolveURLQueryPathInt64Field(r, "parent_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	childGroupID, err := service.ResolveURLQueryPathInt64Field(r, "child_group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	userAllowSubgroups, err := user.AllowSubgroups()
	if err == database.ErrUserNotFound || !userAllowSubgroups {
		return service.InsufficientAccessRightsError
	}

	apiErr := service.NoError

	err = srv.Store.InTransaction(func(s *database.DataStore) error {
		var errInTransaction error
		groupStore := s.Groups()

		var groupData []struct {
			ID   int64  `gorm:"column:ID"`
			Type string `gorm:"column:sType"`
		}

		service.MustNotBeError(groupStore.OwnedBy(user).
			WithWriteLock().
			Select("groups.ID, sType").
			Where("groups.ID IN(?, ?)", parentGroupID, childGroupID).
			Scan(&groupData).Error())

		if len(groupData) < 2 {
			apiErr = service.ErrForbidden(errors.New("insufficient access rights"))
			return apiErr.Error // rollback
		}

		for _, groupRow := range groupData {
			if (groupRow.ID == parentGroupID && groupRow.Type == "UserSelf") ||
				(groupRow.ID == childGroupID &&
					map[string]bool{"Root": true, "RootSelf": true, "RootAdmin": true, "UserAdmin": true}[groupRow.Type]) {
				apiErr = service.ErrForbidden(errors.New("insufficient access rights"))
				return apiErr.Error // rollback
			}
		}

		errInTransaction = groupStore.GroupGroups().CreateRelation(parentGroupID, childGroupID)
		if errInTransaction == database.ErrRelationCycle {
			apiErr = service.ErrForbidden(errInTransaction)
		}
		return errInTransaction
	})

	if apiErr != service.NoError {
		return apiErr
	}

	service.MustNotBeError(err)
	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(nil)))

	return service.NoError
}
