package api

import (
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/swaggo/http-swagger"

	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	_ "github.com/France-ioi/AlgoreaBackend/docs"
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

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {
	r := chi.NewRouter()
	base := service.Base{Store: database.NewDataStore(ctx.db), Config: ctx.config}
	r.Group((&items.Service{Base: base}).SetRoutes)
	r.Group((&groups.Service{Base: base}).SetRoutes)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.NotFound(ctx.notFound)
	return r
}
