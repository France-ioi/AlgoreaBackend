package event

import "context"

// NoopDispatcher is a dispatcher that does nothing.
// Used when no event dispatcher is configured.
type NoopDispatcher struct{}

// Dispatch does nothing and returns nil.
func (d *NoopDispatcher) Dispatch(_ context.Context, _ *Event) error {
	return nil
}

// Ensure NoopDispatcher implements Dispatcher.
var _ Dispatcher = (*NoopDispatcher)(nil)
