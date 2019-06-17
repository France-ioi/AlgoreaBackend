// Package items provides API services for items managing
package items

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `items`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserIDMiddleware(&srv.Config.Auth))
	router.Post("/items/", service.AppHandler(srv.addItem).ServeHTTP)
	router.Get("/items/", service.AppHandler(srv.getList).ServeHTTP)
	router.Get("/items/{item_id}", service.AppHandler(srv.getItem).ServeHTTP)
	router.Put("/items/{item_id}", service.AppHandler(srv.updateItem).ServeHTTP)
	router.Get("/items/{item_id}/as-nav-tree", service.AppHandler(srv.getNavigationData).ServeHTTP)
	router.Put("/items/{item_id}/active-attempt", service.AppHandler(srv.fetchActiveAttempt).ServeHTTP)
	router.Put("/attempts/{groups_attempt_id}/active", service.AppHandler(srv.updateActiveAttempt).ServeHTTP)
	router.Get("/items/{item_id}/attempts", service.AppHandler(srv.getAttempts).ServeHTTP)
	router.Put("/items/{item_id}/strings/{language_id}", service.AppHandler(srv.updateItemString).ServeHTTP)
	router.Post("/items/ask_hint", service.AppHandler(srv.askHint).ServeHTTP)
	router.Post("/items/save-grade", service.AppHandler(srv.saveGrade).ServeHTTP)
}
