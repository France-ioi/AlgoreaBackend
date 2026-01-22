//go:build !prod && !unit

package testhelpers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cucumber/godog"

	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
)

// AnEventShouldHaveBeenDispatched checks that an event of the given type was dispatched.
func (ctx *TestContext) AnEventShouldHaveBeenDispatched(eventType string) error {
	events := ctx.mockEventDispatcher.GetEvents()
	for _, e := range events {
		if e.Type == eventType {
			return nil
		}
	}
	return fmt.Errorf("expected event %q to be dispatched, but it was not. Dispatched events: %v", eventType, eventTypes(events))
}

// AnEventShouldHaveBeenDispatchedWith checks that an event of the given type was dispatched with the specified payload fields.
func (ctx *TestContext) AnEventShouldHaveBeenDispatchedWith(eventType string, jsonPayload *godog.DocString) error {
	events := ctx.mockEventDispatcher.GetEvents()

	data := ctx.preprocessString(jsonPayload.Content)

	expectedPayload := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &expectedPayload); err != nil {
		return fmt.Errorf("failed to parse expected payload JSON: %w", err)
	}

	for _, e := range events {
		if e.Type == eventType {
			if matchesPayload(e.Payload, expectedPayload) {
				return nil
			}
		}
	}

	return fmt.Errorf("expected event %q with payload %v to be dispatched, but found: %v",
		eventType, expectedPayload, eventsOfType(events, eventType))
}

// NoEventShouldHaveBeenDispatched checks that no events were dispatched.
func (ctx *TestContext) NoEventShouldHaveBeenDispatched() error {
	events := ctx.mockEventDispatcher.GetEvents()
	if len(events) > 0 {
		return fmt.Errorf("expected no events to be dispatched, but found: %v", eventTypes(events))
	}
	return nil
}

// NoEventOfTypeShouldHaveBeenDispatched checks that no event of the given type was dispatched.
func (ctx *TestContext) NoEventOfTypeShouldHaveBeenDispatched(eventType string) error {
	events := ctx.mockEventDispatcher.GetEvents()
	for _, e := range events {
		if e.Type == eventType {
			return fmt.Errorf("expected event %q to NOT be dispatched, but it was", eventType)
		}
	}
	return nil
}

func eventTypes(events []event.Event) []string {
	types := make([]string, len(events))
	for i, e := range events {
		types[i] = e.Type
	}
	return types
}

func eventsOfType(events []event.Event, eventType string) []map[string]interface{} {
	var result []map[string]interface{}
	for _, e := range events {
		if e.Type == eventType {
			result = append(result, e.Payload)
		}
	}
	return result
}

func matchesPayload(actual, expected map[string]interface{}) bool {
	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			return false
		}
		if !valuesEqual(actualValue, expectedValue) {
			return false
		}
	}
	return true
}

func valuesEqual(actual, expected interface{}) bool {
	// Handle numeric comparisons (int64 vs float64 etc.)
	actualNum, actualIsNum := toFloat64(actual)
	expectedNum, expectedIsNum := toFloat64(expected)
	if actualIsNum && expectedIsNum {
		return actualNum == expectedNum
	}
	return reflect.DeepEqual(actual, expected)
}

func toFloat64(value interface{}) (float64, bool) {
	switch numValue := value.(type) {
	case int:
		return float64(numValue), true
	case int64:
		return float64(numValue), true
	case float64:
		return numValue, true
	case float32:
		return float64(numValue), true
	}
	return 0, false
}
