// Package threads provides API services for threads managing.
package threads

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `items`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/threads", service.AppHandler(srv.listThreads).ServeHTTP)
	router.Get("/items/{item_id}/participant/{participant_id}/thread", service.AppHandler(srv.getThread).ServeHTTP)
	router.Put("/items/{item_id}/participant/{participant_id}/thread", service.AppHandler(srv.updateThread).ServeHTTP)
}
