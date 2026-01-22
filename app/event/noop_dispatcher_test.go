package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopDispatcher_ReturnsNil(t *testing.T) {
	dispatcher := &NoopDispatcher{}
	event := &Event{Type: "test", Payload: map[string]interface{}{"key": "value"}}

	err := dispatcher.Dispatch(context.Background(), event)

	assert.NoError(t, err)
}

func TestNoopDispatcher_ImplementsDispatcher(_ *testing.T) {
	var _ Dispatcher = (*NoopDispatcher)(nil)
}
