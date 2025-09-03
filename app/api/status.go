package api

import (
	"net/http"
)

func (ctx *Ctx) status(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	// do not use too many internal libs to test the raw server
	_, _ = responseWriter.Write([]byte("The web service is responding! "))
	db := ctx.service.GetStore(httpRequest)
	if db == nil || db.DB == nil {
		_, _ = responseWriter.Write([]byte("The database connection fails."))
	} else {
		_, _ = responseWriter.Write([]byte("The database connection is established."))
	}
}
