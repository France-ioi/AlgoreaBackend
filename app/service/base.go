// Package service provides utilities used for implementing services.
package service

import (
	"net/http"

	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

// Base is the common service context data.
type Base struct {
	store        *database.DataStore
	ServerConfig *viper.Viper
	AuthConfig   *viper.Viper
	DomainConfig []domain.ConfigItem
	TokenConfig  *token.Config
}

// SetGlobalStore sets the global store shared by all the request (should be called only once on start).
func (srv *Base) SetGlobalStore(store *database.DataStore) {
	srv.store = store
}

// GetUser returns the authenticated user data from context.
func (srv *Base) GetUser(r *http.Request) *database.User {
	return auth.UserFromContext(r.Context())
}

// GetSessionID returns the session ID from the request's context.
func (srv *Base) GetSessionID(r *http.Request) int64 {
	return auth.SessionIDFromContext(r.Context())
}

// GetStore returns a data store with the given request's context.
func (srv *Base) GetStore(r *http.Request) *database.DataStore {
	if srv.store == nil {
		return nil
	}
	return database.NewDataStoreWithContext(r.Context(), srv.store.DB)
}

// GetPropagationEndpoint returns the propagation endpoint from the config.
func (srv *Base) GetPropagationEndpoint() string {
	return srv.ServerConfig.GetString("propagation_endpoint")
}
