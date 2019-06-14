package doc

// These definitions are unused by code, just used to generate documentation

// Generic response for errors.
// Possible errors are: 400 (invalid request), 403 (forbidden), 404 (not found), 422 (unprocessable entity), 500 (internal server error)
//
// swagger:response error
type errorResponse struct {
	// in: body
	Body struct {
		// false
		// required: true
		Success bool `json:"success"`
		// Error description, match the HTTP error code (see https://golang.org/src/net/http/status.go)
		// required: true
		Message string `json:"message"`
		// The error message, for developers, only to be used for debugging.
		ErrorText string `json:"error_text,omitempty"`
		// In case of input data validation error, this may contain a map with, as key, the field in error and, as value, an array of strings describing errors in English.
		Errors interface{} `json:"errors,omitempty"`
	}
}

// The request has successfully updated the object
// swagger:response updated
type updated struct {
	// in: body
	Body struct {
		// "updated"
		// enum: updated
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
	}
}
