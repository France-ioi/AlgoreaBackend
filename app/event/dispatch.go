package event

import (
	"context"
	"time"

	"github.com/go-chi/chi/middleware"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// dispatcherConfig holds configuration passed to the Dispatch function.
type dispatcherConfig struct {
	instance string
}

type configContextKeyType int

const configContextKey configContextKeyType = iota

// ContextWithConfig returns a context with the dispatcher configuration.
func ContextWithConfig(ctx context.Context, instance string) context.Context {
	return context.WithValue(ctx, configContextKey, &dispatcherConfig{instance: instance})
}

func configFromContext(ctx context.Context) *dispatcherConfig {
	cfg, _ := ctx.Value(configContextKey).(*dispatcherConfig)
	return cfg
}

// Dispatch sends an event using the dispatcher from context.
// If no dispatcher is configured, this is a silent no-op.
// Errors are logged but not returned (best-effort dispatch).
func Dispatch(ctx context.Context, eventType string, payload map[string]interface{}) {
	dispatcher := DispatcherFromContext(ctx)
	if dispatcher == nil {
		return // silent no-op when no dispatcher configured
	}

	var instance string
	if cfg := configFromContext(ctx); cfg != nil {
		instance = cfg.instance
	}

	evt := &Event{
		Version:   EventVersion,
		Type:      eventType,
		SourceApp: SourceApp,
		Instance:  instance,
		Time:      time.Now(),
		RequestID: middleware.GetReqID(ctx),
		Payload:   payload,
	}

	if err := dispatcher.Dispatch(ctx, evt); err != nil {
		logging.EntryFromContext(ctx).
			WithError(err).
			WithField("event_type", eventType).
			Error("failed to dispatch event")
	}
}
