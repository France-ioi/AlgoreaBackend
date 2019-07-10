package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus" //nolint:depguard
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func TestNew_Success(t *testing.T) {
	assert := assertlib.New(t)
	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)
	assert.NotNil(app.Config)
	assert.NotNil(app.Database)
	assert.NotNil(app.HTTPHandler)
	assert.Len(app.HTTPHandler.Middlewares(), 6)
	assert.True(len(app.HTTPHandler.Routes()) > 0)
}

func TestNew_DBErr(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	patch := monkey.Patch(database.Open, func(interface{}) (*database.DB, error) {
		return nil, errors.New("db opening error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)
	logMsg := hook.LastEntry()
	assert.Equal(logrus.ErrorLevel, logMsg.Level)
	assert.Equal("db opening error", logMsg.Message)
	assert.Equal("database", logMsg.Data["module"])
}

func TestNew_APIErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(api.NewCtx,
		func(conf *config.Root, db *database.DB, tokenConfig *token.Config) (*api.Ctx, error) {
			return nil, errors.New("api creation error")
		})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "api creation error")
}

func TestNew_TokenErr(t *testing.T) {
	assert := assertlib.New(t)
	patch := monkey.Patch(token.Initialize, func(*config.Token) (*token.Config, error) {
		return nil, errors.New("keys loading error")
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(app)
	assert.EqualError(err, "keys loading error")
}

// The goal of the following `TestMiddlewares*` tests are not to test the middleware themselves
// but their interaction (impacted by the order of definition)

func TestMiddlewares_OnPanic(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	app, _ := New()
	router := app.HTTPHandler
	router.Get("/dummy", func(http.ResponseWriter, *http.Request) {
		panic("error in service")
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	nbLogsBeforeRequest := len(hook.AllEntries())
	request, _ := http.NewRequest("GET", srv.URL+"/dummy", nil)
	request.Header.Set("X-Forwarded-For", "1.1.1.1")
	response, _ := http.DefaultClient.Do(request)
	respBody, _ := ioutil.ReadAll(response.Body)

	// check that the error has been handled by the recover
	assert.Equal(http.StatusInternalServerError, response.StatusCode)
	assert.Equal("Internal Server Error\n", string(respBody))
	assert.Equal("text/plain; charset=utf-8", response.Header.Get("Content-type"))
	allLogs := hook.AllEntries()
	assert.Equal(2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the recovere put the error info in the logs
	assert.Equal("error in service", hook.LastEntry().Data["panic"])
	assert.NotNil(hook.LastEntry().Data["stack"])
	// check that the real IP is used in the logs
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
}

func TestMiddlewares_OnSuccess(t *testing.T) {
	assert := assertlib.New(t)
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	app, _ := New()
	router := app.HTTPHandler
	router.Get("/dummy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"data\":\"datadatadata\"}"))
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	nbLogsBeforeRequest := len(hook.AllEntries())
	request, _ := http.NewRequest("GET", srv.URL+"/dummy", nil)
	request.Header.Set("X-Real-IP", "1.1.1.1")
	request.Header.Set("Accept-Encoding", "gzip, deflate")
	response, _ := http.DefaultClient.Do(request)
	assert.NotNil(response.Header.Get("Content-type"))
	assert.Equal("application/json", response.Header.Get("Content-Type"))
	allLogs := hook.AllEntries()
	assert.Equal(2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the real IP is used in the logs
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal("1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
	// check that the compression has been applied but the length in the logs is not altered by compression i
	assert.Equal(23, hook.LastEntry().Data["resp_bytes_length"])
	assert.Equal("gzip", response.Header.Get("Content-Encoding"))
}

func TestNew_MountsPprofInDev(t *testing.T) {
	assert := assertlib.New(t)

	monkey.Patch(appenv.IsEnvDev, func() bool { return true })
	defer monkey.UnpatchAll()

	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest("GET", srv.URL+"/debug", nil)
	response, _ := http.DefaultClient.Do(request)
	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Contains(string(body), "Types of profiles available:")
}

func TestNew_DoesNotMountPprofInEnvironmentsOtherThanDev(t *testing.T) {
	assert := assertlib.New(t)

	monkey.Patch(appenv.IsEnvDev, func() bool { return false })
	defer monkey.UnpatchAll()

	app, err := New()
	assert.NotNil(app)
	assert.NoError(err)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest("GET", srv.URL+"/debug", nil)
	response, _ := http.DefaultClient.Do(request)
	assert.Equal(502, response.StatusCode)
	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Equal("", string(body))
}
