package api

import (
	"net/http"
)

func (ctx *Ctx) status(w http.ResponseWriter, r *http.Request) {
	// do not use too many internal libs to test the raw server
	_, _ = w.Write([]byte("The web service is responding! "))
	db := ctx.service.GetStore(r)
	if db == nil || db.DB == nil {
		_, _ = w.Write([]byte("The database connection fails."))
	} else {
		_, _ = w.Write([]byte("The database connection is established."))
	}
}
