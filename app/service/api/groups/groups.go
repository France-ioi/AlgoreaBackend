package groups

import (
	"github.com/go-chi/chi"
)

type GroupsStore interface {
	GetAll(dest interface{}) error
}

type GroupsService struct {
	Store GroupsStore
}

// New creates a service context
func New(store GroupsStore) *GroupsService {
	return &GroupsService{store}
}

// Router returns the router to the services
func (srv *GroupsService) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", srv.getAll)
	return r
}
