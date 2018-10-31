package api

import (
	"database/sql"
	"net/http/httputil"
	"net/url"

	"github.com/France-ioi/AlgoreaBackend/app/config"

	"github.com/France-ioi/AlgoreaBackend/app/service/api/groups"
	"github.com/go-chi/chi"
)

// Ctx is the context of the root of the API
type Ctx struct {
	config       *config.Root
	db           *sql.DB
	reverseProxy *httputil.ReverseProxy
}

// NewCtx creates a API context
func NewCtx(config *config.Root, db *sql.DB) (*Ctx, error) {
	proxyURL, err := url.Parse(config.ReverseProxy.Server)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	return &Ctx{config, db, proxy}, nil
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Mount("/groups", groups.NewCtx(ctx.db).Router())
	r.NotFound(ctx.notFound)
	return r
}
