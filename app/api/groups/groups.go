package groups

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	s "github.com/France-ioi/AlgoreaBackend/app/service"

	"github.com/go-chi/chi"
)

// Service is the mount point for services related to `groups`
type Service struct {
	Store *database.DataStore
}

// New creates a service context
func New(store *database.DataStore) *Service {
	return &Service{store}
}

// AppendRoutes adds the routes of this pakcage to the parent router
func (srv *Service) AppendRoutes(router *chi.Mux) {
	router.Get("/groups/", s.AppHandler(srv.getAll).ServeHTTP)
}
