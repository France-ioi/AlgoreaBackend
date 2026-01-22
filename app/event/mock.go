package event

import (
	"context"
	"sync"
)

// MockDispatcher is a dispatcher that records events for testing.
type MockDispatcher struct {
	events []Event
	mu     sync.Mutex
}

// NewMockDispatcher creates a new mock dispatcher.
func NewMockDispatcher() *MockDispatcher {
	return &MockDispatcher{}
}

// Dispatch records the event for later inspection.
func (m *MockDispatcher) Dispatch(_ context.Context, event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, *event)
	return nil
}

// GetEvents returns all recorded events.
func (m *MockDispatcher) GetEvents() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Event, len(m.events))
	copy(result, m.events)
	return result
}

// Clear removes all recorded events.
func (m *MockDispatcher) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

// Ensure MockDispatcher implements Dispatcher.
var _ Dispatcher = (*MockDispatcher)(nil)
