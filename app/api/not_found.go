package api

import (
	"net/http"
)

func (ctx *Ctx) notFound(w http.ResponseWriter, _ *http.Request) {
	// to be done in a near future
	w.WriteHeader(404)
}
