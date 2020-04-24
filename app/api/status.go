package api

import (
	"net/http"
)

func (ctx *Ctx) status(w http.ResponseWriter, r *http.Request) {
	// do not use too many internal libs to test the raw server
	_, _ = w.Write([]byte("The web service is responding! ")) // nolint
	if ctx.service.Store == nil || ctx.service.Store.DB == nil {
		_, _ = w.Write([]byte("The database connection fails.")) // nolint
	} else {
		_, _ = w.Write([]byte("The database connection is established.")) // nolint
	}
}
