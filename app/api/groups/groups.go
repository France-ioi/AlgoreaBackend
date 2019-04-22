package groups

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `groups`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserIDMiddleware(&srv.Config.Auth))
	router.Get("/groups/", service.AppHandler(srv.getAll).ServeHTTP)
	router.Get("/groups/{group_id}/recent_activity", service.AppHandler(srv.getRecentActivity).ServeHTTP)
	router.Get("/groups/{group_id}", service.AppHandler(srv.getGroup).ServeHTTP)
	router.Put("/groups/{group_id}", service.AppHandler(srv.updateGroup).ServeHTTP)
	router.Post("/groups/{group_id}/change_password", service.AppHandler(srv.changePassword).ServeHTTP)
	router.Get("/groups/{group_id}/children", service.AppHandler(srv.getChildren).ServeHTTP)
	router.Get("/groups/{group_id}/requests", service.AppHandler(srv.getRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/accept_requests", service.AppHandler(srv.acceptRequests).ServeHTTP)
	router.Post("/groups/{parent_group_id}/reject_requests", service.AppHandler(srv.rejectRequests).ServeHTTP)

	router.Post("/group-relations/{parent_group_id}/{child_group_id}", service.AppHandler(srv.addChild).ServeHTTP)
	router.Delete("/group-relations/{parent_group_id}/{child_group_id}", service.AppHandler(srv.removeChild).ServeHTTP)
}

func checkThatUserOwnsTheGroup(store *database.DataStore, user *database.User, groupID int64) service.APIError {
	var count int64
	if err := store.GroupAncestors().OwnedByUser(user).
		Where("idGroupChild = ?", groupID).Count(&count).Error(); err != nil {
		if err == database.ErrUserNotFound {
			return service.InsufficientAccessRightsError
		}
		return service.ErrUnexpected(err)
	}
	if count == 0 {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func checkThatUserHasRightsForDirectRelation(store *database.DataStore, user *database.User,
	parentGroupID, childGroupID int64) service.APIError {
	groupStore := store.Groups()

	var groupData []struct {
		ID   int64  `gorm:"column:ID"`
		Type string `gorm:"column:sType"`
	}

	err := groupStore.OwnedBy(user).
		WithWriteLock().
		Select("groups.ID, sType").
		Where("groups.ID IN(?, ?)", parentGroupID, childGroupID).
		Scan(&groupData).Error()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if len(groupData) < 2 {
		return service.InsufficientAccessRightsError
	}

	for _, groupRow := range groupData {
		if (groupRow.ID == parentGroupID && groupRow.Type == "UserSelf") ||
			(groupRow.ID == childGroupID &&
				map[string]bool{"Root": true, "RootSelf": true, "RootAdmin": true, "UserAdmin": true}[groupRow.Type]) {
			return service.InsufficientAccessRightsError
		}
	}
	return service.NoError
}
