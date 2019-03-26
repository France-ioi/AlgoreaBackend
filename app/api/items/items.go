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
	router.Get("/items/{itemID}", service.AppHandler(srv.getItem).ServeHTTP)
	router.Get("/items/{itemID}/as-nav-tree", service.AppHandler(srv.getNavigationData).ServeHTTP)
}
