package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	s "github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/go-chi/chi"
)

// ItemService is the mount point for services related to `items`
type ItemService struct {
	Store *database.DataStore
}

// New creates a service context
func New(store *database.DataStore) *ItemService {
	return &ItemService{store}
}

// Router returns the router to the services
func (srv *ItemService) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", s.AppHandler(srv.addItem).ServeHTTP)
	return r
}
