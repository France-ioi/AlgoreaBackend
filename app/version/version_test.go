package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddVersionHeader(t *testing.T) {
	expectedVersion := "myversion-1.2.3"
	oldVersion := version
	version = expectedVersion
	defer func() { version = oldVersion }()

	middleware := AddVersionHeader
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("dummy body"))
	})
	recorder := httptest.NewRecorder()
	middleware(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "dummy body", recorder.Body.String())

	assert.Equal(t, expectedVersion, recorder.Header().Get("Backend-Version"))
}
