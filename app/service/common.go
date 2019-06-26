package service

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Base is the common service context data
type Base struct {
	Store       *database.DataStore
	Config      *config.Root
	TokenConfig *token.Config
}

// GetUser returns the authenticated user data from context
func (srv *Base) GetUser(r *http.Request) *database.User {
	return auth.UserFromContext(r.Context(), srv.Store.Users())
}

// SetQueryLimit limits the number of records of the given query according to the `limit` request parameter
// The default limit is 500
// The optional `limits` argument can be either [default_limit] or [default_limit, max_possible_limit]
func SetQueryLimit(r *http.Request, db *database.DB, limits ...int) *database.DB {
	limit, err := ResolveURLQueryGetInt64Field(r, "limit")
	defaultLimitToUse := int64(500)
	maxPossibleLimit := int64(1000)
	if len(limits) > 0 {
		defaultLimitToUse = int64(limits[0])
	}
	if len(limits) > 1 {
		maxPossibleLimit = int64(limits[1])
	}
	if err != nil || limit < 0 {
		limit = defaultLimitToUse
	}
	if limit > maxPossibleLimit {
		limit = maxPossibleLimit
	}
	return db.Limit(limit)
}
