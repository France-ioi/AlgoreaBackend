package doc

// These definitions are unused by code, just used to generate documentation

// Use the error responses below if you want to provide a generic error for a service.
// Otherwise, use errors (see `errors.go`)
//
// Example:
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"

// BadRequest. There is an error in the input data which was provided (path or body) to this service.
// swagger:response badRequestPOSTPUTPATCHResponse
type badRequestPOSTPUTPATCHResponse struct {
	// in: body
	Body struct{ badRequest }
}

// Unauthorized. The authorization token has not been provided or is invalid.
// swagger:response unauthorizedResponse
type unauthorizedResponse struct {
	// in: body
	Body struct{ unauthorized }
}

// Forbidden. One of the permission requirements (cfr service description) is not met.
// (note that some permission error may end up in not-found or bad-request errors)
// swagger:response forbiddenResponse
type forbiddenResponse struct {
	// in: body
	Body struct{ forbidden }
}

// Internal Error. An unexpected error has happened on the server (e.g., uncaught database error).
// If the problem persists, it should be reported.
// swagger:response internalErrorResponse
type internalErrorResponse struct {
	// in: body
	Body struct{ internalError }
}
