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
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Ctx is the context of the root of the API
type Ctx struct {
	service      *service.Base
	db           *database.DB
	ServerConfig *viper.Viper
	AuthConfig   *viper.Viper
	DomainConfig []domain.AppConfigItem
	TokenConfig  *token.Config
}

// NewCtx creates a API context
func NewCtx(db *database.DB, serverConfig, authConfig *viper.Viper, domainConfig []domain.AppConfigItem, tokenConfig *token.Config) *Ctx {
	return &Ctx{nil, db, serverConfig, authConfig, domainConfig, tokenConfig}
}

// SetAuthConfig update the auth config used by the API
func (ctx *Ctx) SetAuthConfig(authConfig *viper.Viper) {
	ctx.AuthConfig = authConfig
	ctx.service.AuthConfig = authConfig
}

// Router provides routes for the whole API
func (ctx *Ctx) Router() *chi.Mux {

	r := chi.NewRouter()
	ctx.service = &service.Base{
		Store:        database.NewDataStore(ctx.db),
		ServerConfig: ctx.ServerConfig,
		AuthConfig:   ctx.AuthConfig,
		DomainConfig: ctx.DomainConfig,
		TokenConfig:  ctx.TokenConfig,
	}
	r.Group((&auth.Service{Base: ctx.service}).SetRoutes)
	r.Group((&contests.Service{Base: ctx.service}).SetRoutes)
	r.Group((&items.Service{Base: ctx.service}).SetRoutes)
	r.Group((&groups.Service{Base: ctx.service}).SetRoutes)
	r.Group((&answers.Service{Base: ctx.service}).SetRoutes)
	r.Group((&currentuser.Service{Base: ctx.service}).SetRoutes)
	r.Get("/status", ctx.status)
	r.NotFound(service.NotFound)
	return r
}
