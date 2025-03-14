// Package contests provides API services for contests managing.
package contests

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `contests`
// swagger:ignore
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route contests.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/contests/administered", service.AppHandler(srv.getAdministeredList).ServeHTTP)

	router.Get("/contests/{item_id}/groups/{group_id}/additional-times",
		service.AppHandler(srv.getGroupAdditionalTimes).ServeHTTP)
}
