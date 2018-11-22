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

// Router returns the router to the services
func (srv *Service) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", s.AppHandler(srv.getAll).ServeHTTP)
	return r
}
