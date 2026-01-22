package event

import "context"

// Dispatcher defines the interface for sending events to external systems.
type Dispatcher interface {
	// Dispatch sends an event. Returns an error if the dispatch fails.
	Dispatch(ctx context.Context, event *Event) error
}
