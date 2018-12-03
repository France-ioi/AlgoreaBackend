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

// APIError represents an error as returned by this application. It works in
// tandem with AppHandler for easy handling of errors.
type APIError struct {
  HTTPStatusCode int
  Error          error
}

// NoError is an APIError to be returned when there is no error
var NoError = APIError{0, nil}

func (e APIError) httpResponse() render.Renderer {
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
func (e APIError) Render(w http.ResponseWriter, r *http.Request) error {
  render.Status(r, e.HTTPStatusCode)
  return nil
}

// ErrInvalidRequest is for errors caused by invalid request input
// It results in a 400 Invalid request response
func ErrInvalidRequest(err error) APIError {
  return APIError{http.StatusBadRequest, err}
}

// ErrForbidden is for errors caused by a lack of permissions globally or on a requested object
// It results in a 403 Invalid request response
func ErrForbidden(err error) APIError {
  return APIError{http.StatusForbidden, err}
}

// ErrUnexpected is for internal errors (not supposed to fail) not directly caused by the user input
// It results in a 500 Internal Server Error response
func ErrUnexpected(err error) APIError {
  return APIError{http.StatusInternalServerError, err}
}
