package service

import (
	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"net/http"
)

// Base is the common service context data
type Base struct {
	Store  *database.DataStore
	Config *config.Root
}

func (srv *Base) GetUser(r *http.Request) *auth.User {
	return auth.UserFromContext(r.Context(), srv.Store.Users())
}
