package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// AppHandler is a type that implements http.Handler and makes handling
// errors easier. When its method returns an error, it prints it to the logs
// and shows a JSON formatted error to the user.
type AppHandler func(http.ResponseWriter, *http.Request) APIError

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiErr := NoError
	var shouldLogError bool
	var errorToLog string
	defer func() {
		if p := recover(); p != nil {
			switch err := p.(type) {
			case APIError:
				apiErr = err
			case error:
				if errors.Is(err, context.DeadlineExceeded) {
					apiErr = ErrRequestTimeout()
				} else {
					apiErr = ErrUnexpected(fmt.Errorf("unknown error"))
					shouldLogError = true
					errorToLog = err.Error()
				}
			default:
				apiErr = ErrUnexpected(fmt.Errorf("unknown error"))
				errorToLog = fmt.Sprintf("%+v", err)
				shouldLogError = true
			}
			if shouldLogError {
				logging.GetLogEntry(r).Errorf("unexpected error: %s, stack trace: %s", errorToLog, debug.Stack())
			}
		}
		if apiErr != NoError { // apiErr is an APIError, not builtin.error
			_ = render.Render(w, r, apiErr.httpResponse()) // nolint, never fails
		}
	}()
	apiErr = fn(w, r)
}
