package service

import (
	"net/http"

	"github.com/go-chi/render"
)

// AppHandler is a type that implements http.Handler and makes handling
// errors easier. When its method returns an error, it prints it to the logs
// and shows a JSON formatted error to the user.
type AppHandler func(http.ResponseWriter, *http.Request) *AppError

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil { // err is *AppError, not os.Error
		if err2 := render.Render(w, r, err.httpResponse()); err2 != nil {
			panic(err2) // if unable to render errors, panic
		}
	}
}
