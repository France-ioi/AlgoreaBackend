package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

func TestDispatch_NoopWhenNoDispatcher(_ *testing.T) {
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	// Should not panic even without a dispatcher
	Dispatch(ctx, TypeSubmissionCreated, map[string]interface{}{"test": "value"})
}

func TestDispatch_SendsEventToDispatcher(t *testing.T) {
	mockDispatcher := NewMockDispatcher()
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	ctx = ContextWithDispatcher(ctx, mockDispatcher)

	payload := map[string]interface{}{
		"author_id": int64(123),
		"item_id":   int64(456),
	}

	Dispatch(ctx, TypeSubmissionCreated, payload)

	events := mockDispatcher.GetEvents()
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, TypeSubmissionCreated, event.Type)
	assert.Equal(t, SourceApp, event.SourceApp)
	assert.Equal(t, EventVersion, event.Version)
	assert.Equal(t, payload, event.Payload)
	assert.NotZero(t, event.Time)
}

func TestDispatch_IncludesInstanceFromConfig(t *testing.T) {
	mockDispatcher := NewMockDispatcher()
	ctx, _, _ := logging.NewContextWithNewMockLogger()
	ctx = ContextWithDispatcher(ctx, mockDispatcher)
	ctx = ContextWithConfig(ctx, "prod")

	Dispatch(ctx, TypeSubmissionCreated, map[string]interface{}{})

	events := mockDispatcher.GetEvents()
	require.Len(t, events, 1)
	assert.Equal(t, "prod", events[0].Instance)
}

func TestDispatch_LogsErrorOnFailure(t *testing.T) {
	ctx, _, hook := logging.NewContextWithNewMockLogger()
	failingDispatcher := &failingDispatcher{}
	ctx = ContextWithDispatcher(ctx, failingDispatcher)

	Dispatch(ctx, TypeSubmissionCreated, map[string]interface{}{})

	// Check that an error was logged
	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, "failed to dispatch event", hook.LastEntry().Message)
}

// failingDispatcher is a test dispatcher that always returns an error.
type failingDispatcher struct{}

func (d *failingDispatcher) Dispatch(_ context.Context, _ *Event) error {
	return assert.AnError
}
