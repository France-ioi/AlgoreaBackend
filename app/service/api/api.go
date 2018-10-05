package api

import (
	"database/sql"

	"github.com/France-ioi/AlgoreaBackend/app/service/api/groups"
	"github.com/go-chi/chi"
)

// Ctx is the context of the root of the API
type Ctx struct {
	db       *sql.DB
	groupCtx *groups.Ctx
}

// NewCtx creates a API context
func NewCtx(db *sql.DB) *Ctx {
	return &Ctx{db, groups.NewCtx(db)}
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Mount("/groups", ctx.groupCtx.Router())
	return r
}
