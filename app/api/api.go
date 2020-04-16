package api

import (
	"github.com/go-chi/chi"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app/api/answers"
	"github.com/France-ioi/AlgoreaBackend/app/api/auth"
	"github.com/France-ioi/AlgoreaBackend/app/api/contests"
	"github.com/France-ioi/AlgoreaBackend/app/api/currentuser"
	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Ctx is the context of the root of the API
type Ctx struct {
	db           *database.DB
	serverConfig *viper.Viper
	domainConfig *viper.Viper
	authConfig   *viper.Viper
	tokenConfig  *token.Config
}

// NewCtx creates a API context
func NewCtx(db *database.DB, serverConfig, domainConfig, authConfig *viper.Viper, tokenConfig *token.Config) *Ctx {
	return &Ctx{db, serverConfig, domainConfig, authConfig, tokenConfig}
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {

	r := chi.NewRouter()
	base := service.Base{
		Store:        database.NewDataStore(ctx.db),
		ServerConfig: ctx.serverConfig,
		DomainConfig: ctx.domainConfig,
		AuthConfig:   ctx.authConfig,
		TokenConfig:  ctx.tokenConfig,
	}
	r.Group((&auth.Service{Base: base}).SetRoutes)
	r.Group((&contests.Service{Base: base}).SetRoutes)
	r.Group((&items.Service{Base: base}).SetRoutes)
	r.Group((&groups.Service{Base: base}).SetRoutes)
	r.Group((&answers.Service{Base: base}).SetRoutes)
	r.Group((&currentuser.Service{Base: base}).SetRoutes)
	r.Get("/status", ctx.status)
	r.NotFound(service.NotFound)
	return r
}
