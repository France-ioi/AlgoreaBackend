// Package event provides utilities for dispatching domain events.
package event

import "time"

// Event represents a domain event that can be dispatched to external systems.
type Event struct {
	Version   string                 `json:"version"`              // Event schema version (e.g., "1.0")
	Type      string                 `json:"type"`                 // Event type (e.g., "submission_created")
	SourceApp string                 `json:"source_app"`           // Source application (always "algoreabackend")
	Instance  string                 `json:"instance,omitempty"`   // Optional instance identifier (e.g., "prod", "staging")
	Time      time.Time              `json:"time"`                 // When the event occurred
	RequestID string                 `json:"request_id,omitempty"` // Request ID for correlation
	Payload   map[string]interface{} `json:"payload"`              // Event-specific data
}

const (
	// SourceApp is the static source application identifier.
	SourceApp = "algoreabackend"

	// EventVersion is the current event schema version.
	// Increment minor for non-breaking changes (adding optional fields).
	// Increment major for breaking changes (removing fields, changing semantics).
	EventVersion = "1.2"
)
