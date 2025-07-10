// Package api contains the web services API.
package api

import (
	"github.com/go-chi/chi"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api/answers"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/currentuser"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/groups"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/items"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/threads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/api/users"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

// Ctx is the context of the root of the API.
type Ctx struct {
	service *service.Base
}

// Router provides routes for the whole API.
func Router(db *database.DB, serverConfig, authConfig *viper.Viper, domainConfig []domain.ConfigItem,
	tokenConfig *token.Config,
) (*Ctx, *chi.Mux) {
	router := chi.NewRouter()

	srv := &service.Base{
		ServerConfig: serverConfig,
		AuthConfig:   authConfig,
		DomainConfig: domainConfig,
		TokenConfig:  tokenConfig,
	}
	srv.SetGlobalStore(database.NewDataStore(db))

	ctx := &Ctx{srv}
	router.Group((&auth.Service{Base: srv}).SetRoutes)
	router.Group((&items.Service{Base: srv}).SetRoutes)
	router.Group((&threads.Service{Base: srv}).SetRoutes)
	router.Group((&groups.Service{Base: srv}).SetRoutes)
	router.Group((&answers.Service{Base: srv}).SetRoutes)
	router.Group((&currentuser.Service{Base: srv}).SetRoutes)
	router.Group((&users.Service{Base: srv}).SetRoutes)
	router.Get("/status", ctx.status)
	router.NotFound(service.NotFound)

	return ctx, router
}
