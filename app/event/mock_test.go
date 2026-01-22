package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockDispatcher_RecordsEvents(t *testing.T) {
	mock := NewMockDispatcher()

	event1 := &Event{Type: "event1", Payload: map[string]interface{}{"key1": "value1"}}
	event2 := &Event{Type: "event2", Payload: map[string]interface{}{"key2": "value2"}}

	err := mock.Dispatch(context.Background(), event1)
	require.NoError(t, err)

	err = mock.Dispatch(context.Background(), event2)
	require.NoError(t, err)

	events := mock.GetEvents()
	require.Len(t, events, 2)
	assert.Equal(t, "event1", events[0].Type)
	assert.Equal(t, "event2", events[1].Type)
}

func TestMockDispatcher_Clear(t *testing.T) {
	mock := NewMockDispatcher()

	event1 := &Event{Type: "event1"}
	_ = mock.Dispatch(context.Background(), event1)

	require.Len(t, mock.GetEvents(), 1)

	mock.Clear()

	assert.Empty(t, mock.GetEvents())
}

func TestMockDispatcher_GetEventsReturnsCopy(t *testing.T) {
	mock := NewMockDispatcher()

	event1 := &Event{Type: "event1"}
	_ = mock.Dispatch(context.Background(), event1)

	events := mock.GetEvents()
	events[0].Type = "modified"

	// Original should not be modified
	originalEvents := mock.GetEvents()
	assert.Equal(t, "event1", originalEvents[0].Type)
}
