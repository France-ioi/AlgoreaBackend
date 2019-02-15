package answers

import (
	"github.com/go-chi/chi"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `answers`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route answers
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(auth.UserIDMiddleware(&srv.Config.Auth))
	router.Get("/answers", service.AppHandler(srv.getAnswers).ServeHTTP)
}
