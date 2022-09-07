// Package auth provides API services related to authentication
package auth

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `auth`
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Post("/auth/temp-user", service.AppHandler(srv.createTempUser).ServeHTTP)

	router.With(middleware.AllowContentType("", "application/json", "application/x-www-form-urlencoded")).
		Post("/auth/token", service.AppHandler(srv.createAccessToken).ServeHTTP)
	router.With(auth.UserMiddleware(srv.Base)).
		Post("/auth/logout", service.AppHandler(srv.logout).ServeHTTP)
}
