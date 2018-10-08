package service

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render generates the HTTP response from ErrResponse
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// stdError generates a response with standard error message and HTTP code, and no extra debug message.
// Should be used for basic expected errors which does not require extra debugging or explanation
// Find codes in `net/http/status.go`
func stdError(code int) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: code,
		StatusText:     http.StatusText(code),
	}
}

// detailedError generated an error response from a HTTP code (from `net/http/status.go`) and a given error for debugging purposes
func detailedError(code int, err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: code,
		StatusText:     http.StatusText(code),
		ErrorText:      err.Error(),
	}
}

// ErrNotFound is a 404 Not Found response
var ErrNotFound = stdError(http.StatusNotFound)

// ErrInvalidRequest generates a 400 Invalid request response
func ErrInvalidRequest(err error) render.Renderer {
	return detailedError(http.StatusBadRequest, err)
}

// ErrServer generates a 500 Internal Server Error response
// Use this for errors not caused by the user input (not supposed to fail)
func ErrServer(err error) render.Renderer {
	return detailedError(http.StatusInternalServerError, err)
}
