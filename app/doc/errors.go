package doc

// These definitions are unused by code, just used to generate documentation

// Use the errors below if you want to provide a custom explanation for an error type (error code) for a service.
// Otherwise, use error responses. (see `error_responses.go`)
//
// Example:
//   "400":
//     description: Bad Request. The id you have provided
//     schema:
//       "$ref": "#/definitions/badRequest"

type genericError struct {
	// false
	// example: false
	// required: true
	Success bool `json:"success"`
	// Error description, match the HTTP error code (see https://golang.org/src/net/http/status.go)
	// required: true
	Message string `json:"message"`
	// The error message, for developers, only to be used for debugging.
	ErrorText string `json:"error_text,omitempty"`
}

type badRequest struct {
	genericError
	// required: true
	// enum: Bad Request
	Message string `json:"message"`
	// In case of input data validation error, this may contain a map with, as key, the field in error
	// and, as value, an array of strings describing errors in English.
	Errors interface{} `json:"errors,omitempty"`
}

type unauthorized struct {
	genericError
	// required: true
	// enum: Unauthorized
	Message string `json:"message"`
}

type forbidden struct {
	genericError
	// required: true
	// enum: Forbidden
	Message string `json:"message"`
}

type notFound struct {
	genericError
	// required: true
	// enum: Not Found
	Message string `json:"message"`
}

type unprocessableEntity struct {
	genericError
	// required: true
	// enum: Unprocessable Entity
	Message string `json:"message"`
}

type internalError struct {
	genericError
	// required: true
	// enum: Internal Server Error
	Message string `json:"message"`
}
