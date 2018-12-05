package groups

import (
  "github.com/France-ioi/AlgoreaBackend/app/auth"
  s "github.com/France-ioi/AlgoreaBackend/app/service"

  "github.com/go-chi/chi"
)

// Service is the mount point for services related to `groups`
type Service struct {
  s.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
  router.Use(auth.UserIDMiddleware(&srv.Config.Auth))
  router.Get("/groups/", s.AppHandler(srv.getAll).ServeHTTP)
}
