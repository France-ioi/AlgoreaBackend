package service

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/database/users"
)

// Base is the common service context data
type Base struct {
	Store  *database.DataStore
	Config *config.Root
}

// GetUser returns the authenticated user data from context
func (srv *Base) GetUser(r *http.Request) *auth.User {
	return auth.UserFromContext(r.Context(), users.NewStore(srv.Store))
}

// SetQueryLimit limits the number of records of the given query according to the `limit` request parameter
// The default limit is 500
func (srv *Base) SetQueryLimit(r *http.Request, db database.DB) database.DB {
	limit, err := ResolveURLQueryGetInt64Field(r, "limit")
	if err != nil || limit < 0 {
		limit = 500
	}
	return db.Limit(limit)
}
