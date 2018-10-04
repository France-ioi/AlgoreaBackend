package app

import (
	"github.com/go-chi/chi"
)

// InitRouter creates, inits and returns the GIN router
func InitRouter() *chi.Mux {

	r := chi.NewRouter()

	return r
}
