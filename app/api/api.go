package api

import (
	"net/http/httputil"
	"net/url"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"

	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/api/items"
	"github.com/go-chi/chi"
)

// Ctx is the context of the root of the API
type Ctx struct {
	config       *config.Root
	db           *database.DB
	reverseProxy *httputil.ReverseProxy
}

// NewCtx creates a API context
func NewCtx(config *config.Root, db *database.DB) (*Ctx, error) {
	var err error
	var proxyURL *url.URL

	if proxyURL, err = url.Parse(config.ReverseProxy.Server); err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	return &Ctx{config, db, proxy}, nil
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	dataStore := database.NewDataStore(ctx.db)
	items.New(dataStore).AppendRoutes(r)
	groups.New(dataStore).AppendRoutes(r)
	r.NotFound(ctx.notFound)
	return r
}
