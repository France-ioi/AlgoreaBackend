package api

import (
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"

	"github.com/France-ioi/AlgoreaBackend/app/api/answers"
	"github.com/France-ioi/AlgoreaBackend/app/api/auth"
	"github.com/France-ioi/AlgoreaBackend/app/api/currentuser"
	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Ctx is the context of the root of the API
type Ctx struct {
	config       *config.Root
	db           *database.DB
	reverseProxy *httputil.ReverseProxy
	tokenConfig  *token.Config
}

// NewCtx creates a API context
func NewCtx(conf *config.Root, db *database.DB, tokenConfig *token.Config) (*Ctx, error) {
	var err error
	var proxyURL *url.URL

	if proxyURL, err = url.Parse(conf.ReverseProxy.Server); err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	return &Ctx{config: conf, db: db, reverseProxy: proxy, tokenConfig: tokenConfig}, nil
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	base := service.Base{
		Store:       database.NewDataStore(ctx.db),
		Config:      ctx.config,
		TokenConfig: ctx.tokenConfig,
	}
	r.Group((&auth.Service{Base: base}).SetRoutes)
	r.Group((&items.Service{Base: base}).SetRoutes)
	r.Group((&groups.Service{Base: base}).SetRoutes)
	r.Group((&answers.Service{Base: base}).SetRoutes)
	r.Group((&currentuser.Service{Base: base}).SetRoutes)
	r.Get("/status", ctx.status)
	r.NotFound(ctx.notFound)
	return r
}
