// Package api contains the web services API.
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
	"github.com/France-ioi/AlgoreaBackend/app/api/threads"
	"github.com/France-ioi/AlgoreaBackend/app/api/users"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Ctx is the context of the root of the API
type Ctx struct {
	service *service.Base
}

// Router provides routes for the whole API
func Router(db *database.DB, serverConfig, authConfig *viper.Viper, domainConfig []domain.ConfigItem,
	tokenConfig *token.Config) (*Ctx, *chi.Mux) {
	r := chi.NewRouter()

	srv := &service.Base{
		ServerConfig: serverConfig,
		AuthConfig:   authConfig,
		DomainConfig: domainConfig,
		TokenConfig:  tokenConfig,
	}
	srv.SetGlobalStore(database.NewDataStore(db))

	ctx := &Ctx{srv}
	r.Group((&auth.Service{Base: srv}).SetRoutes)
	r.Group((&contests.Service{Base: srv}).SetRoutes)
	r.Group((&items.Service{Base: srv}).SetRoutes)
	r.Group((&threads.Service{Base: srv}).SetRoutes)
	r.Group((&groups.Service{Base: srv}).SetRoutes)
	r.Group((&answers.Service{Base: srv}).SetRoutes)
	r.Group((&currentuser.Service{Base: srv}).SetRoutes)
	r.Group((&users.Service{Base: srv}).SetRoutes)
	r.Get("/status", ctx.status)
	r.NotFound(service.NotFound)

	return ctx, r
}
