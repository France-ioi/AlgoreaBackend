package service

import (
	"net/http"
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

// MaxSelectExecutionTime caps read-only SELECTs issued while handling the request.
func MaxSelectExecutionTime(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			ctx := database.ContextWithMaxSelectExecutionTime(httpRequest.Context(), d)
			next.ServeHTTP(responseWriter, httpRequest.WithContext(ctx))
		})
	}
}
