package app

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus" //nolint:depguard
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/version"
)

/* note that the tests of app.New() are very incomplete (even if all exec path are covered) */

func TestNew_Success(t *testing.T) {
	appenv.SetDefaultEnvToTest()
	t.Setenv("ALGOREA_SERVER__COMPRESS", "1")
	app, err := New()
	assert.NotNil(t, app)
	require.NoError(t, err)
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.Database)
	assert.NotNil(t, app.HTTPHandler)
	assert.NotNil(t, app.apiCtx)
	assert.Len(t, app.HTTPHandler.Middlewares(), 9)
	require.NotEmpty(t, app.HTTPHandler.Routes())
	assert.Equal(t, "/*", app.HTTPHandler.Routes()[0].Pattern) // test default val
}

func TestNew_SuccessNoCompress(t *testing.T) {
	appenv.SetDefaultEnvToTest()
	t.Setenv("ALGOREA_SERVER__COMPRESS", "false")
	app, _ := New()
	assert.Len(t, app.HTTPHandler.Middlewares(), 8)
}

func TestNew_NotDefaultRootPath(t *testing.T) {
	appenv.SetDefaultEnvToTest()
	t.Setenv("ALGOREA_SERVER__ROOTPATH", "/api")
	app, err := New()
	require.NoError(t, err)
	require.NotEmpty(t, app.HTTPHandler.Routes())
	assert.Equal(t, "/api/*", app.HTTPHandler.Routes()[0].Pattern)
}

func TestNew_DBErr(t *testing.T) {
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	expectedError := errors.New("db opening error")
	patch := monkey.Patch(database.Open, func(interface{}) (*database.DB, error) {
		return nil, expectedError
	})
	defer patch.Unpatch()
	app, err := New()
	assert.Nil(t, app)
	require.Equal(t, expectedError, err)
	logMsg := hook.LastEntry()
	require.NotNil(t, logMsg)
	assert.Equal(t, logrus.ErrorLevel, logMsg.Level)
	assert.Equal(t, "db opening error", logMsg.Message)
	assert.Equal(t, "database", logMsg.Data["module"])
}

func TestNew_RandSeedingFailed(t *testing.T) {
	expectedError := errors.New("some error")
	patch := monkey.Patch(crand.Read, func([]byte) (int, error) {
		return 1, expectedError
	})
	defer patch.Unpatch()
	assert.PanicsWithValue(t, "cannot seed the randomizer", func() { _, _ = New() })
}

func TestNew_DBConfigError(t *testing.T) {
	patch := monkey.Patch(DBConfig, func(_ *viper.Viper) (config *mysql.Config, err error) {
		return nil, errors.New("dberror")
	})
	defer patch.Unpatch()
	_, err := New()
	assert.EqualError(t, err, "unable to load the 'database' configuration: dberror")
}

func TestNew_TokenConfigError(t *testing.T) {
	patch := monkey.Patch(LoadConfig, func() *viper.Viper {
		globalConfig := viper.New()
		globalConfig.Set("token.PublicKeyFile", "notafile")
		return globalConfig
	})
	defer patch.Unpatch()
	_, err := New()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestNew_DomainsConfigError(t *testing.T) {
	patch := monkey.Patch(LoadConfig, func() *viper.Viper {
		globalConfig := viper.New()
		globalConfig.Set("domains", []int{1, 2})
		return globalConfig
	})
	defer patch.Unpatch()
	_, err := New()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load the 'domain' configuration: 2 error(s) decoding")
}

// The goal of the following `TestMiddlewares*` tests are not to test the middleware themselves
// but their interaction (impacted by the order of definition)

func TestMiddlewares_OnPanic(t *testing.T) {
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
	request, _ := http.NewRequest(http.MethodGet, srv.URL+"/dummy", http.NoBody)
	request.Header.Set("X-Forwarded-For", "1.1.1.1")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	if err != nil {
		return
	}
	respBody, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()

	// check that the error has been handled by the recover
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Equal(t, "Internal Server Error\n", string(respBody))
	assert.Equal(t, "text/plain; charset=utf-8", response.Header.Get("Content-Type"))
	allLogs := hook.AllEntries()
	require.Equal(t, 2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(t, allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the recover put the error info in the logs
	assert.Equal(t, "error in service", hook.LastEntry().Data["panic"])
	assert.NotNil(t, hook.LastEntry().Data["stack"])
	// check that the real IP is used in the logs
	assert.Equal(t, "1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal(t, "1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
}

func TestMiddlewares_OnSuccess(t *testing.T) {
	t.Setenv("ALGOREA_SERVER__COMPRESS", "1")
	hook, restoreFct := logging.MockSharedLoggerHook()
	defer restoreFct()
	app, _ := New()
	router := app.HTTPHandler
	router.Get("/dummy", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"data\":\"datadatadata\"}"))
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	nbLogsBeforeRequest := len(hook.AllEntries())
	request, _ := http.NewRequest(http.MethodGet, srv.URL+"/dummy", http.NoBody)
	request.Header.Set("X-Real-IP", "1.1.1.1")
	request.Header.Set("Accept-Encoding", "gzip, deflate")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	if err != nil {
		return
	}
	defer func() { _ = response.Body.Close() }()
	assert.NotNil(t, response.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	allLogs := hook.AllEntries()
	require.Equal(t, 2, len(allLogs)-nbLogsBeforeRequest)
	// check that the req id is correct
	assert.Equal(t, allLogs[len(allLogs)-1].Data["req_id"], allLogs[len(allLogs)-2].Data["req_id"])
	// check that the real IP is used in the logs
	assert.Equal(t, "1.1.1.1", allLogs[len(allLogs)-1].Data["remote_addr"])
	assert.Equal(t, "1.1.1.1", allLogs[len(allLogs)-2].Data["remote_addr"])
	// check that the compression has been applied but the length in the logs is not altered by compression i
	assert.Equal(t, 23, hook.LastEntry().Data["resp_bytes_length"])
	assert.Equal(t, "gzip", response.Header.Get("Content-Encoding"))
	assert.Equal(t, version.Version, response.Header.Get("Backend-Version"))
}

func TestNew_MountsPprofInDev(t *testing.T) {
	appenv.SetDefaultEnvToTest()
	monkey.Patch(appenv.IsEnvDev, func() bool { return true })
	defer monkey.UnpatchAll()

	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest(http.MethodGet, srv.URL+"/debug", http.NoBody)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Types of profiles available:")
}

func TestNew_DoesNotMountPprofInEnvironmentsOtherThanDev(t *testing.T) {
	monkey.Patch(appenv.IsEnvDev, func() bool { return false })
	defer monkey.UnpatchAll()

	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app)

	srv := httptest.NewServer(app.HTTPHandler)
	defer srv.Close()

	request, _ := http.NewRequest(http.MethodGet, srv.URL+"/debug", http.NoBody)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	assert.Equal(t, 404, response.StatusCode)
}

func TestNew_DisableResultsPropagation(t *testing.T) {
	for _, configSettingValue := range []bool{true, false} {
		configSettingValue := configSettingValue
		t.Run(fmt.Sprintf("disableResultsPropagation=%t", configSettingValue), func(t *testing.T) {
			t.Setenv("ALGOREA_SERVER__DISABLERESULTSPROPAGATION", fmt.Sprintf("%t", configSettingValue))
			app, _ := New()
			assert.Equal(t, configSettingValue, database.NewDataStore(app.Database).IsResultsPropagationProhibited())

			router := app.HTTPHandler
			router.Get("/dummy", func(_ http.ResponseWriter, r *http.Request) {
				assert.Equal(t, configSettingValue, database.NewDataStoreWithContext(r.Context(), app.Database).IsResultsPropagationProhibited())
			})

			srv := httptest.NewServer(router)
			defer srv.Close()

			request, _ := http.NewRequest(http.MethodGet, srv.URL+"/dummy", http.NoBody)
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			_, _ = io.ReadAll(response.Body)
			_ = response.Body.Close()
		})
	}
}
