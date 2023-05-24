// Package version exposes the app version
package version

import "net/http"

// Version is the app version.
var Version string

// AddVersionHeader adds the Backend-Version header in the responses.
func AddVersionHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Backend-Version", Version)

		next.ServeHTTP(w, r)
	})
}
