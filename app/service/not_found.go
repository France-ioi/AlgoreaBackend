package service

import (
	"net/http"

	"github.com/go-chi/render"
)

// NotFound is a basic HTTP handler which generates a 404 error.
func NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, ErrNotFound(nil).httpResponse()) // never fails
}
