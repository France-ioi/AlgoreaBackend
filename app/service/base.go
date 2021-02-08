package service

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/token"

	"github.com/spf13/viper"
)

// Base is the common service context data
type Base struct {
	Store        *database.DataStore
	ServerConfig *viper.Viper
	AuthConfig   *viper.Viper
	DomainConfig []domain.ConfigItem
	TokenConfig  *token.Config
}

// GetUser returns the authenticated user data from context
func (srv *Base) GetUser(r *http.Request) *database.User {
	return auth.UserFromContext(r.Context())
}
