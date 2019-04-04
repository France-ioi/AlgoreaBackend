package groups

import (
	"errors"

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
}

func checkThatUserOwnsTheGroup(store *database.DataStore, user *database.User, groupID int64) service.APIError {
	var count int64
	service.MustNotBeError(
		store.GroupAncestors().OwnedByUser(user).
			Where("idGroupChild = ?", groupID).Count(&count).Error())
	if count == 0 {
		return service.ErrForbidden(errors.New("insufficient access rights"))
	}
	return service.NoError
}
