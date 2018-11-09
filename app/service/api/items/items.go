package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/go-chi/chi"
)

type ItemService struct {
	Store *database.DataStore
}

// New creates a service context
func New(store *database.DataStore) *ItemService {
	return &ItemService{store}
}

// Router returns the router to the services
func (ctx *ItemService) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", ctx.addItem)
	return r
}
