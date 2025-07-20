// Package version exposes the app version
package version

import "net/http"

var version = "unknown"

// AddVersionHeader adds the Backend-Version header in the responses.
func AddVersionHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Backend-Version", GetVersion())

		next.ServeHTTP(w, r)
	})
}

// GetVersion returns the version of the application.
func GetVersion() string {
	return version
}
