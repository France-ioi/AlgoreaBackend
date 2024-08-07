// Package users provides API services for users managing.
package users

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `users`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/users/{user_id}", service.AppHandler(srv.getUser).ServeHTTP)
	router.Get("/users/by-login/{login}", service.AppHandler(srv.getUser).ServeHTTP)
	router.Post("/users/{target_user_id}/generate-profile-edit-token", service.AppHandler(srv.generateProfileEditToken).ServeHTTP)
}
