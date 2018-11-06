package groups

import (
	"github.com/go-chi/chi"
)

type GroupStore interface {
	GetAll(dest interface{}) error
}

type GroupsService struct {
	Store GroupStore
}

// New creates a service context
func New(store GroupStore) *GroupsService {
	return &GroupsService{store}
}

// Router returns the router to the services
func (srv *GroupsService) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", srv.getAll)
	return r
}
