package service

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

// ErrorResponse is an extension of the response for returning errors
type ErrorResponse struct {
	Response
	ErrorText string      `json:"error_text,omitempty"` // application-level error message, for debugging
	Errors    interface{} `json:"errors,omitempty"`     // form errors
}

// APIError represents an error as returned by this application. It works in
// tandem with AppHandler for easy handling of errors.
type APIError struct {
	HTTPStatusCode int
	Error          error
}

// NoError is an APIError to be returned when there is no error
var NoError = APIError{0, nil}

var InsufficientAccessRightsError = ErrForbidden(errors.New("insufficient access rights"))

func (e APIError) httpResponse() render.Renderer {
	response := Response{
		HTTPStatusCode: e.HTTPStatusCode,
		Success:        false,
		Message:        http.StatusText(e.HTTPStatusCode),
	}
	result := ErrorResponse{Response: response}
	if e.Error == nil {
		return &result
	}

	if fieldErrors, ok := e.Error.(FieldErrors); ok {
		result.Errors = fieldErrors
	}

	result.ErrorText = e.Error.Error() // FIXME: should be disabled in prod
	if len(result.ErrorText) > 0 {
		result.ErrorText = strings.ToUpper(result.ErrorText[0:1]) + result.ErrorText[1:]
	}

	return &result
}

// ErrInvalidRequest is for errors caused by invalid request input
// It results in a 400 Invalid request response
func ErrInvalidRequest(err error) APIError {
	return APIError{http.StatusBadRequest, err}
}

// ErrForbidden is for errors caused by a lack of permissions globally or on a requested object
// It results in a 403 Forbidden
func ErrForbidden(err error) APIError {
	return APIError{http.StatusForbidden, err}
}

// ErrNotFound is for errors caused by absence of a requested object
// It results in a 404 Not Found
func ErrNotFound(err error) APIError {
	return APIError{http.StatusNotFound, err}
}

// ErrUnexpected is for internal errors (not supposed to fail) not directly caused by the user input
// It results in a 500 Internal Server Error response
func ErrUnexpected(err error) APIError {
	return APIError{http.StatusInternalServerError, err}
}

// MustNotBeError panics if the error is not nil
func MustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
