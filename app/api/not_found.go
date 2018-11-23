package api

import (
  "net/http"
)

func (ctx *Ctx) notFound(w http.ResponseWriter, r *http.Request) {
  ctx.reverseProxy.ServeHTTP(w, r)
}
