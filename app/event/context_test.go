package event

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatcherFromContext_ReturnsNilWhenNotSet(t *testing.T) {
	ctx := context.Background()
	dispatcher := DispatcherFromContext(ctx)
	assert.Nil(t, dispatcher)
}

func TestDispatcherFromContext_ReturnsDispatcher(t *testing.T) {
	mockDispatcher := NewMockDispatcher()
	ctx := ContextWithDispatcher(context.Background(), mockDispatcher)
	retrieved := DispatcherFromContext(ctx)
	assert.Equal(t, mockDispatcher, retrieved)
}

func TestContextWithDispatcher_PreservesExistingContext(t *testing.T) {
	type testKey string
	const key testKey = "test"
	const value = "value"

	parentCtx := context.WithValue(context.Background(), key, value)
	mockDispatcher := NewMockDispatcher()

	ctx := ContextWithDispatcher(parentCtx, mockDispatcher)

	assert.Equal(t, value, ctx.Value(key))
	assert.Equal(t, mockDispatcher, DispatcherFromContext(ctx))
}

func TestContextWithDispatcherMiddleware(t *testing.T) {
	mockDispatcher := NewMockDispatcher()
	middleware := ContextWithDispatcherMiddleware(mockDispatcher)

	var called bool
	var retrievedDispatcher Dispatcher

	handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		retrievedDispatcher = DispatcherFromContext(r.Context())
	}))

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.True(t, called)
	assert.Equal(t, mockDispatcher, retrievedDispatcher)
}

func TestContextWithDispatcherMiddleware_PreservesExistingContext(t *testing.T) {
	type testKey string
	const key testKey = "test"
	const value = "value"

	mockDispatcher := NewMockDispatcher()
	middleware := ContextWithDispatcherMiddleware(mockDispatcher)

	var retrievedValue interface{}

	handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		retrievedValue = r.Context().Value(key)
	}))

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	req = req.WithContext(context.WithValue(req.Context(), key, value))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, value, retrievedValue)
}

func TestContextWithDispatcherMiddleware_DoesNotOverwriteExistingDispatcher(t *testing.T) {
	existingDispatcher := NewMockDispatcher()
	newDispatcher := NewMockDispatcher()

	middleware := ContextWithDispatcherMiddleware(newDispatcher)

	var retrievedDispatcher Dispatcher

	handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		retrievedDispatcher = DispatcherFromContext(r.Context())
	}))

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)
	// Pre-set dispatcher in context (simulates test injection)
	req = req.WithContext(ContextWithDispatcher(req.Context(), existingDispatcher))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	// Should get the existing dispatcher, not the new one
	assert.Equal(t, existingDispatcher, retrievedDispatcher)
}
