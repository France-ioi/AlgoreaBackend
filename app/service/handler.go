package service

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// AppHandler is a type that implements http.Handler and makes handling
// errors easier. When its method returns an error, it prints it to the logs
// and shows a JSON formatted error to the user.
type AppHandler func(http.ResponseWriter, *http.Request) APIError

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiErr := NoError
	defer func() {
		if p := recover(); p != nil {
			switch err := p.(type) {
			case error:
				apiErr = ErrUnexpected(err)
			default:
				apiErr = ErrUnexpected(fmt.Errorf("unknown error: %+v", err))
			}
			logging.GetLogEntry(r).Errorf("unexpected error: %s", apiErr.Error)
		}
		if apiErr != NoError { // apiErr is an APIError, not os.Error
			_ = render.Render(w, r, apiErr.httpResponse()) // nolint, never fails
		}
	}()
	apiErr = fn(w, r)
}
