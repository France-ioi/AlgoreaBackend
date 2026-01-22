package event

import (
	"context"
	"net/http"
)

type dispatcherContextKeyType int

const dispatcherContextKey dispatcherContextKeyType = iota

// DispatcherFromContext returns the dispatcher from the given context.
// Returns nil if no dispatcher is configured.
func DispatcherFromContext(ctx context.Context) Dispatcher {
	d, _ := ctx.Value(dispatcherContextKey).(Dispatcher)
	return d
}

// ContextWithDispatcher returns a copy of the given context with the dispatcher set.
func ContextWithDispatcher(ctx context.Context, d Dispatcher) context.Context {
	return context.WithValue(ctx, dispatcherContextKey, d)
}

// ContextWithDispatcherMiddleware returns a middleware that sets the dispatcher in the request context.
// If a dispatcher is already set in the context (e.g., for testing), it will not be overwritten.
func ContextWithDispatcherMiddleware(dispatcher Dispatcher) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Don't overwrite if a dispatcher is already set (allows test injection)
			if DispatcherFromContext(request.Context()) != nil {
				next.ServeHTTP(responseWriter, request)
				return
			}
			next.ServeHTTP(responseWriter, request.WithContext(ContextWithDispatcher(request.Context(), dispatcher)))
		})
	}
}
