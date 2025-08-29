//go:build !unit

package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"bou.ke/monkey"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiAuth "github.com/France-ioi/AlgoreaBackend/v2/app/api/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestService_refreshAccessToken_NotAllowRefreshTokenRaces(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timePatch := testhelpers.MockDBTime("2025-08-27 22:32:01")
	defer testhelpers.RestoreDBTime(timePatch)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		access_tokens: [{token: accesstoken, session_id: 1, expires_at: "2025-08-27 22:42:01", issued_at: "2025-08-27 22:22:01"}]
		sessions: [{session_id: 1, user_id: 2, refresh_token: "firstrefreshtoken"}]
	`)
	defer func() { _ = db.Close() }()

	expectedClientID := "1234"
	expectedClientSecret := "secret"
	loginModuleStubServer := createLoginModuleStubServer(expectedClientID, expectedClientSecret)
	done := make(chan bool, 1)
	lockCalledChan := make(chan struct{}, 1)
	lockCalledByTestedCodeChan := make(chan struct{}, 1)
	expectedLockName := fmt.Sprintf("session_%d", auth.MockCtxSessionID)
	var funcCalled atomic.Bool
	doRequest := func(withTimeout bool) {
		var cancelFunc context.CancelFunc
		var patchGuard *monkey.PatchGuard
		patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
			func(store *database.DataStore, lockName string, timeout time.Duration, funcToCall func(store *database.DataStore) error) error {
				lockCalledByTestedCodeChan <- struct{}{} // notify that the code under test called WithNamedLock()
				assert.Equal(t, expectedLockName, lockName)

				if withTimeout {
					cancelFunc()
					ctx := store.GetContext()
					(*valueCtxInterface)(unsafe.Pointer(&ctx)).p.timerCtx.err = context.DeadlineExceeded //nolint:gosec // imitate a timeout
				}

				patchGuard.Unpatch()
				defer patchGuard.Restore()
				newFuncToCall := func(store *database.DataStore) error {
					funcCalled.Store(true)
					return funcToCall(store)
				}
				return store.WithNamedLock(lockName, timeout, newFuncToCall)
			})
		defer patchGuard.Unpatch()

		base := service.Base{}
		base.SetGlobalStore(database.NewDataStore(db))
		router := chi.Router(chi.NewRouter())
		logger, logHook := logging.NewMockLogger()
		router.Use(logging.ContextWithLoggerMiddleware(logger))
		router.Use(auth.MockUserMiddleware(&database.User{GroupID: 2}))
		router.Use(middleware.RequestLogger(&logging.StructuredLogger{}))

		srv := &Service{Base: &base}
		srv.AuthConfig = viper.New()
		srv.AuthConfig.Set("loginModuleURL", loginModuleStubServer.URL)
		srv.AuthConfig.Set("clientID", expectedClientID)
		srv.AuthConfig.Set("clientSecret", expectedClientSecret)
		routerForAddingHandlers := router
		if withTimeout {
			routerForAddingHandlers = routerForAddingHandlers.With(func(next http.Handler) http.Handler {
				handlerFunc := func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
					var ctx context.Context
					ctx, cancelFunc = context.WithDeadline(httpRequest.Context(), time.Now().Add(1*time.Hour))
					defer cancelFunc()

					httpRequest = httpRequest.WithContext(ctx)
					next.ServeHTTP(responseWriter, httpRequest)
				}
				return http.HandlerFunc(handlerFunc)
			})
		}

		routerForAddingHandlers.
			With(func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					r = r.WithContext(contextWithParsedRequestData(r.Context(), map[string]interface{}{}))
					next.ServeHTTP(w, r)
				}
				return http.HandlerFunc(fn)
			}).Post("/auth/token", service.AppHandler(srv.refreshAccessToken).ServeHTTP)

		request, err := http.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(""))
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		logs := (&loggingtest.Hook{Hook: logHook}).GetAllLogs()

		if withTimeout {
			assert.Equal(t, 408, recorder.Code)
			assert.Contains(t, logs, "context deadline exceeded")
		} else {
			assert.Equal(t, 201, recorder.Code)
			body := recorder.Body.String()
			assert.JSONEq(t,
				`{"success":true,"message":"created","data":{"access_token":"newaccesstoken","expires_in":78901234}}`,
				body)
		}
		done <- true // notify that doRequest() has finished
	}

	store := database.NewDataStore(db)

	// lock the session
	go func() {
		assert.NoError(t, store.WithNamedLock(expectedLockName, 1*time.Second, func(_ *database.DataStore) error {
			lockCalledChan <- struct{}{} // notify that the lock has been acquired by the test
			<-lockCalledByTestedCodeChan // wait for the code under test to call WithNamedLock()
			return nil                   // unlock letting doRequest() to perform
		}))
	}()
	<-lockCalledChan // wait for the concurrent test function to acquire the lock
	doRequest(false)
	<-done // wait until the service finishes
	assert.True(t, funcCalled.Load())

	funcCalled.Store(false)
	require.NoError(t, store.AccessTokens().Delete("token = 'newaccesstoken'").Error())

	// check that the service timeouts if the session is locked for too long

	// lock the session
	go func() {
		assert.NoError(t, store.WithNamedLock(expectedLockName, 1*time.Second, func(_ *database.DataStore) error {
			lockCalledChan <- struct{}{} // notify that the lock has been acquired by the test
			<-done                       // wait until the service finishes with timeout
			return nil                   // unlock
		}))
	}()

	<-lockCalledChan // wait for the concurrent test function to acquire the lock
	doRequest(true)
	<-lockCalledByTestedCodeChan // be sure the code under test has called WithNamedLock()
	assert.False(t, funcCalled.Load())
}

func createLoginModuleStubServer(expectedClientID, expectedClientSecret string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		_ = httpRequest.ParseForm()
		if httpRequest.URL.Path == "/oauth/token" &&
			httpRequest.Form.Get("grant_type") == "refresh_token" &&
			httpRequest.Form.Get("client_id") == expectedClientID &&
			httpRequest.Form.Get("client_secret") == expectedClientSecret {
			responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
			responseWriter.WriteHeader(http.StatusOK)
			_, _ = responseWriter.Write([]byte(
				`{"access_token": "newaccesstoken", "refresh_token": "new` + httpRequest.Form.Get("refresh_token") +
					`", "expires_in": 78901234}`))
		}
	}))
}

type cancelCtx struct {
	context.Context //nolint:containedctx // it is not us who store the context in the structure
	_               sync.Mutex
	_               atomic.Value
	_               map[interface{}]struct{}
	err             error
}

type timerCtx struct {
	*cancelCtx
}

type valueCtx struct {
	_        unsafe.Pointer
	timerCtx *timerCtx
}
type valueCtxInterface struct {
	t unsafe.Pointer
	p *valueCtx
}

type Service struct {
	*service.Base
}

func (srv *Service) refreshAccessToken(w http.ResponseWriter, r *http.Request) error {
	return srvRefreshAccessToken((*apiAuth.Service)(srv), w, r)
}

//go:linkname srvRefreshAccessToken github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.(*Service).refreshAccessToken
func srvRefreshAccessToken(srv *apiAuth.Service, w http.ResponseWriter, r *http.Request) error

//go:linkname contextWithParsedRequestData github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.contextWithParsedRequestData
func contextWithParsedRequestData(ctx context.Context, requestData map[string]interface{}) context.Context
