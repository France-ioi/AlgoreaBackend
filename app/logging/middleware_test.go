package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_Success(t *testing.T) {
	logger, hook := NewMockLogger()

	doRequest(logger, false)

	assert.Len(t, hook.AllEntries(), 3)

	// First entry: request started
	entryData := hook.AllEntries()[0].Data
	checkCommon(t, entryData)
	assert.Equal(t, "request started", hook.AllEntries()[0].Message)

	// Second entry: in-service message
	entryData = hook.AllEntries()[1].Data
	checkCommon(t, entryData)
	assert.Equal(t, "in service log", hook.AllEntries()[1].Message)
	assert.Equal(t, 42, entryData["my_key"])
	assert.Equal(t, 1, entryData["opt_one"])
	assert.Equal(t, "bar", entryData["foo"])

	// Third entry: request complete
	entryData = hook.AllEntries()[2].Data
	checkCommon(t, entryData)
	assert.Equal(t, "request complete", hook.AllEntries()[2].Message)
	assert.Equal(t, 10, entryData["resp_bytes_length"])
	respElapsedMs, ok := entryData["resp_elapsed_ms"].(float64)
	require.True(t, ok)
	assert.Less(t, respElapsedMs, 3000.0, "Expected <3.0s, got: %f", respElapsedMs)
	assert.Equal(t, 200, entryData["resp_status"])
}

func TestMiddleware_Panic(t *testing.T) {
	logger, hook := NewMockLogger()

	doRequest(logger, true)

	assert.Len(t, hook.AllEntries(), 3)

	// Third entry: panic
	entryData := hook.LastEntry().Data
	checkCommon(t, entryData)
	assert.NotNil(t, entryData["stack"])
	assert.Equal(t, "my panic msg", entryData["panic"])
}

func doRequest(logger *Logger, forcePanic bool) {
	// setting up the server with 1 service and using the logger middleware
	loggerMiddleware := NewStructuredLogger()
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		LogEntrySetField(httpRequest, "my_key", 42)
		LogEntrySetFields(httpRequest, map[string]interface{}{"opt_one": 1, "foo": "bar"})
		GetLogEntry(httpRequest).Print("in service log")
		if forcePanic {
			panic("my panic msg")
		}
		responseWriter.WriteHeader(http.StatusOK)
		_, _ = responseWriter.Write([]byte("dummy body"))
	})
	// use the chi `Recoverer` middleware to catch panic and log it
	// use the chi `RequestID` middleware to include request id in it (appear in logs)
	// The order of the middlewares is crucial!
	mainSrv := httptest.NewTLSServer(ContextWithLoggerMiddleware(logger)(
		middleware.RequestID(loggerMiddleware(middleware.Recoverer(handler)))))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest(http.MethodGet, mainSrv.URL+"/a_path", http.NoBody)
	client := mainSrv.Client()
	response, err := client.Do(mainRequest)
	if err == nil {
		_ = response.Body.Close()
	}
}

func checkCommon(t *testing.T, entryData logrus.Fields) {
	t.Helper()

	assert.Equal(t, "web", entryData["type"])
	assert.Equal(t, "https", entryData["http_scheme"])
	assert.Equal(t, "HTTP/1.1", entryData["http_proto"])
	assert.Equal(t, "GET", entryData["http_method"])
	remoteAddr, remoteAddrOK := entryData["remote_addr"].(string)
	require.True(t, remoteAddrOK)
	assert.Regexp(t, "^127.0.0.1:", remoteAddr)
	assert.Equal(t, "Go-http-client/1.1", entryData["user_agent"])
	uri, uriOK := entryData["uri"].(string)
	require.True(t, uriOK)
	assert.Regexp(t, "^https://127.0.0.1:\\d*/a_path$", uri)
	assert.NotNil(t, entryData["req_id"])
}
