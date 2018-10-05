package groups

import (
	"database/sql"

	"github.com/go-chi/chi"
)

// Ctx is the context
type Ctx struct {
	db *sql.DB
}

// NewCtx creates a service context
func NewCtx(db *sql.DB) *Ctx {
	return &Ctx{db}
}

// Router returns the router to the services
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", ctx.getAll)
	return r
}
