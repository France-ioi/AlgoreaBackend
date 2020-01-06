package doc

// These definitions are unused by code, just used to generate documentation

// Use the error responses below if you want to provide a generic error for a service.
// Otherwise, use errors (see `errors.go`)
//
// Example:
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"

// BadRequest. There is an error in the input data which was provided (path or body) to this service.
// swagger:response badRequestResponse
type badRequestResponse struct {
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

// Not Found. The requested object (or its scope) is not found
// (note that this error may be caused by insufficient access rights)
// swagger:response notFoundResponse
type notFoundResponse struct {
	// in: body
	Body struct{ notFound }
}

// Unprocessable Entity. Returned by services performing groups relations transitions to indicate
// that the transition is impossible.
// swagger:response unprocessableEntityResponse
type unprocessableEntityResponse struct {
	// in: body
	Body struct{ unprocessableEntity }
}

// Unprocessable Entity. Returned by services performing groups relations transitions to indicate
// that the transition is impossible because of missing approvals.
// swagger:response unprocessableEntityResponseWithMissingApprovals
type unprocessableEntityResponseWithMissingApprovals struct {
	// in:body
	Body struct {
		unprocessableEntityWithMissingApprovals
	}
}

// Internal Error. An unexpected error has happened on the server (e.g., uncaught database error).
// If the problem persists, it should be reported.
// swagger:response internalErrorResponse
type internalErrorResponse struct {
	// in: body
	Body struct{ internalError }
}
