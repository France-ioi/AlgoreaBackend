package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	assertlib "github.com/stretchr/testify/assert"
)

func TestLogging_Success(t *testing.T) {
	assert := assertlib.New(t)
	logger, hook := test.NewNullLogger()

	doRequest(logger, false)

	assert.Len(hook.AllEntries(), 3)

	// First entry: request started
	entryData := hook.AllEntries()[0].Data
	checkCommon(assert, entryData)
	assert.Equal("request started", hook.AllEntries()[0].Message)

	// Second entry: in-service message
	entryData = hook.AllEntries()[1].Data
	checkCommon(assert, entryData)
	assert.Equal("in service log", hook.AllEntries()[1].Message)
	assert.Equal(42, entryData["my_key"])
	assert.Equal(1, entryData["opt_one"])
	assert.Equal("bar", entryData["foo"])

	// Third entry: request complete
	entryData = hook.AllEntries()[2].Data
	checkCommon(assert, entryData)
	assert.Equal("request complete", hook.AllEntries()[2].Message)
	assert.Equal(10, entryData["resp_bytes_length"])
	assert.True(entryData["resp_elapsed_ms"].(float64) < 0.5, "Expected <0.5s, got: %f", entryData["resp_elapsed_ms"].(float64))
	assert.Equal(200, entryData["resp_status"])
}

func TestLogging_Panic(t *testing.T) {
	assert := assertlib.New(t)
	logger, hook := test.NewNullLogger()

	doRequest(logger, true)

	assert.Len(hook.AllEntries(), 3)

	// Third entry: panic
	entryData := hook.LastEntry().Data
	checkCommon(assert, entryData)
	assert.NotNil(entryData["stack"])
	assert.Equal("my panic msg", entryData["panic"])
}

func doRequest(logger *logrus.Logger, forcePanic bool) {
	// setting up the server with 1 service and using the logger middleware
	loggerMiddleware := NewStructuredLogger(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogEntrySetField(r, "my_key", 42)
		LogEntrySetFields(r, map[string]interface{}{"opt_one": 1, "foo": "bar"})
		GetLogEntry(r).Print("in service log")
		if forcePanic {
			panic("my panic msg")
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("dummy body"))
		}
	})
	// use the chi `Recoverer` middleware to catch panic and log it
	// use the chi `RequestID` middleware to include request id in it (appear in logs)
	// The order of the middlewares is crucial!
	mainSrv := httptest.NewTLSServer(middleware.RequestID(loggerMiddleware(middleware.Recoverer(handler))))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest("GET", mainSrv.URL+"/a_path", nil)
	client := mainSrv.Client()
	client.Do(mainRequest)
}

func checkCommon(assert *assertlib.Assertions, entryData logrus.Fields) {
	assert.NotNil(entryData["ts"])
	assert.Equal("web", entryData["type"])
	assert.Equal("https", entryData["http_scheme"])
	assert.Equal("HTTP/1.1", entryData["http_proto"])
	assert.Equal("GET", entryData["http_method"])
	assert.Regexp("^127.0.0.1:", entryData["remote_addr"].(string))
	assert.Equal("Go-http-client/1.1", entryData["user_agent"])
	assert.Regexp("^https://127.0.0.1:\\d*/a_path$", entryData["uri"].(string))
	assert.NotNil(entryData["req_id"])
}
