// Package cachecontrol provides HTTP middleware that forbids storing API responses.
package cachecontrol

import "net/http"

// AddNoStoreHeader sets Cache-Control: no-store on every response before the next
// handler runs. A handler that sets its own Cache-Control replaces this default.
func AddNoStoreHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
