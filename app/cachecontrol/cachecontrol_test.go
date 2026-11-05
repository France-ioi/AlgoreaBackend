package cachecontrol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddNoStoreHeader(t *testing.T) {
	middleware := AddNoStoreHeader
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("dummy body"))
	})
	recorder := httptest.NewRecorder()
	middleware(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "dummy body", recorder.Body.String())

	assert.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
}

func TestAddNoStoreHeader_HandlerOverrideWins(t *testing.T) {
	middleware := AddNoStoreHeader
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.Header().Set("Cache-Control", "public, max-age=60")
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("cached body"))
	})
	recorder := httptest.NewRecorder()
	middleware(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "cached body", recorder.Body.String())

	assert.Equal(t, "public, max-age=60", recorder.Header().Get("Cache-Control"))
}
