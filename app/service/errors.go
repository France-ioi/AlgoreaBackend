package service

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrorResponse is an extension of the response for returning errors
type ErrorResponse struct {
	Response
	ErrorText string `json:"error_text,omitempty"` // application-level error message, for debugging
}

// AppError represents an error as returned by this application. It works in
// tandem with AppHandler for easy handling of errors.
type AppError struct {
	HTTPStatusCode int
	Error          error
}

func (e *AppError) httpResponse() render.Renderer {
	response := Response{
		HTTPStatusCode: e.HTTPStatusCode,
		Success:        false,
		Message:        http.StatusText(e.HTTPStatusCode),
	}
	if e.Error == nil {
		return &ErrorResponse{Response: response}
	}
	return &ErrorResponse{
		Response:  response,
		ErrorText: e.Error.Error(), // FIXME: should be disabled in prod
	}
}

// Render generates the HTTP response from ErrResponse
func (e *AppError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest is for errors caused by invalid request input
// It results in a 400 Invalid request response
func ErrInvalidRequest(err error) *AppError {
	return &AppError{http.StatusBadRequest, err}
}

// ErrUnexpected is for internal errors (not supposed to fail) not directly caused by the user input
// It results in a 500 Internal Server Error response
func ErrUnexpected(err error) *AppError {
	return &AppError{http.StatusInternalServerError, err}
}
