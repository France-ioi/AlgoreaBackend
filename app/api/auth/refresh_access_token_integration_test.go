//go:build !unit

package auth_test

import (
	"context"
	"errors"
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

const (
	expectedClientID     = "1234"
	expectedClientSecret = "secret"
	expectedRefreshToken = "firstrefreshtoken"
	expectedBody         = `{"success":true,"message":"created","data":{"access_token":"newaccesstoken","expires_in":78901234}}`
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

		recorder, logs, _, err := callRefreshAccessTokenServiceWithLoginModuleStub(withTimeout, db, false, &cancelFunc)
		require.NoError(t, err)

		if withTimeout {
			assert.Equal(t, 408, recorder.Code)
			assert.Contains(t, logs, "context deadline exceeded")
		} else {
			assert.Equal(t, 201, recorder.Code)
			body := recorder.Body.String()
			assert.JSONEq(t, expectedBody, body)
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

func callRefreshAccessTokenServiceWithLoginModuleStub(
	withTimeout bool, db *database.DB, isTempUser bool, cancelFuncToSet *context.CancelFunc,
) (recorder *httptest.ResponseRecorder, logs string, loginModuleCallCount int, err error) {
	loginModuleStubServer, loginModuleCallAtomicCount := createLoginModuleStubServer(expectedRefreshToken)
	defer loginModuleStubServer.Close()

	base := service.Base{}
	base.SetGlobalStore(database.NewDataStore(db))
	router := chi.Router(chi.NewRouter())
	logger, logHook := logging.NewMockLogger()
	router.Use(logging.ContextWithLoggerMiddleware(logger))
	router.Use(auth.MockUserMiddleware(&database.User{GroupID: 2, IsTempUser: isTempUser}))
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
				ctx, cancelFunc := context.WithDeadline(httpRequest.Context(), time.Now().Add(1*time.Hour))
				if cancelFuncToSet != nil {
					*cancelFuncToSet = cancelFunc
				}
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
				r = r.WithContext(contextWithParsedCookieParameters(r.Context(), &apiAuth.CookieParameters{}))
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		}).Post("/auth/token", service.AppHandler(srv.refreshAccessToken).ServeHTTP)

	request, err := http.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(""))
	if err != nil {
		return nil, "", 0, err
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	logs = (&loggingtest.Hook{Hook: logHook}).GetAllLogs()
	return recorder, logs, int(loginModuleCallAtomicCount.Load()), nil
}

func TestService_refreshAccessToken_ConcurrentRequestsGetUpdatedAccessTokenAndSkipsRefreshing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timePatch := testhelpers.MockDBTime("2025-08-27 22:32:01")
	defer testhelpers.RestoreDBTime(timePatch)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		access_tokens: [{token: accesstoken, session_id: 1, expires_at: "2025-08-27 22:42:01", issued_at: "2025-08-27 22:22:01"}]
		sessions: [{session_id: 1, user_id: 2, refresh_token: "firstrefreshtoken"}]
	`)
	defer func() { _ = db.Close() }()

	funcCalledChan := make(chan struct{})
	concurrentRequestCalledChan := make(chan struct{})
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(store *database.DataStore, lockName string, timeout time.Duration, funcToCall func(store *database.DataStore) error) error {
			patchGuard.Unpatch()
			defer patchGuard.Restore()
			newFuncToCall := func(store *database.DataStore) error {
				funcCalledChan <- struct{}{} // notify that the code under test called the func inside WithNamedLock()
				return funcToCall(store)
			}
			return store.WithNamedLock(lockName, timeout, newFuncToCall)
		})
	defer patchGuard.Unpatch()

	go func() {
		<-funcCalledChan // wait until the first request calls the function inside WithNamedLock()
		recorder, _, loginModuleCallCount, err := callRefreshAccessTokenServiceWithLoginModuleStub(false, db, false, nil)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		body := recorder.Body.String()
		assert.Equal(t, 201, recorder.Code)
		assert.Equal(t, 0, loginModuleCallCount) // the login module should not be called for the second time
		assert.JSONEq(t, expectedBody, body)
		concurrentRequestCalledChan <- struct{}{} // notify the first request to continue
	}()
	recorder, _, loginModuleCallCount, err := callRefreshAccessTokenServiceWithLoginModuleStub(false, db, false, nil)
	require.NoError(t, err)
	<-concurrentRequestCalledChan // wait until the concurrent request is called

	assert.Equal(t, 201, recorder.Code)
	body := recorder.Body.String()
	assert.JSONEq(t, expectedBody, body)
	assert.Equal(t, 1, loginModuleCallCount)
}

func TestService_refreshAccessToken_ConcurrentRequestsGetUpdatedAccessTokenAndSkipsRefreshing_TempUser(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	timePatch := testhelpers.MockDBTime("2025-08-27 22:32:01")
	defer testhelpers.RestoreDBTime(timePatch)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		users: [{group_id: 2, temp_user: true}]
		access_tokens: [{token: accesstoken, session_id: 1, expires_at: "2025-08-27 22:42:01", issued_at: "2025-08-27 22:22:01"}]
		sessions: [{session_id: 1, user_id: 2, refresh_token: "firstrefreshtoken"}]
	`)
	defer func() { _ = db.Close() }()

	funcCalledChan := make(chan struct{})
	concurrentRequestCalledChan := make(chan struct{})
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&database.DataStore{}), "WithNamedLock",
		func(store *database.DataStore, lockName string, timeout time.Duration, funcToCall func(store *database.DataStore) error) error {
			patchGuard.Unpatch()
			defer patchGuard.Restore()
			newFuncToCall := func(store *database.DataStore) error {
				funcCalledChan <- struct{}{} // notify that the code under test called the func inside WithNamedLock()
				return funcToCall(store)
			}
			return store.WithNamedLock(lockName, timeout, newFuncToCall)
		})
	var cnt atomic.Uint32
	monkey.Patch(auth.GenerateKey, func() (string, error) {
		newValue := cnt.Add(1)
		if newValue != 1 {
			return "", errors.New("repeated calls to GenerateKey() are not expected")
		}
		return "newaccesstoken", nil
	})
	defer monkey.UnpatchAll()

	const expectedBody = `{"success":true,"message":"created","data":{"access_token":"newaccesstoken","expires_in":7200}}`

	go func() {
		<-funcCalledChan // wait until the first request calls the function inside WithNamedLock()
		recorder, _, loginModuleCallCount, err := callRefreshAccessTokenServiceWithLoginModuleStub(false, db, true, nil)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		body := recorder.Body.String()
		assert.Equal(t, 201, recorder.Code)
		assert.Equal(t, 0, loginModuleCallCount) // the login module should not be called for the second time
		assert.JSONEq(t, expectedBody, body)
		concurrentRequestCalledChan <- struct{}{} // notify the first request to continue
	}()
	recorder, _, loginModuleCallCount, err := callRefreshAccessTokenServiceWithLoginModuleStub(false, db, true, nil)
	require.NoError(t, err)
	select {
	case <-concurrentRequestCalledChan: // wait until the concurrent request is called
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for the concurrent request to be called")
	}

	assert.Equal(t, 201, recorder.Code)
	body := recorder.Body.String()
	assert.JSONEq(t, expectedBody, body)
	assert.Equal(t, 0, loginModuleCallCount)
}

func createLoginModuleStubServer(expectedRefreshToken string) (*httptest.Server, *atomic.Uint32) {
	var loginModuleCalledCount atomic.Uint32
	return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		loginModuleCalledCount.Add(1)
		_ = httpRequest.ParseForm()
		if httpRequest.URL.Path == "/oauth/token" &&
			httpRequest.Form.Get("grant_type") == "refresh_token" &&
			httpRequest.Form.Get("client_id") == expectedClientID &&
			httpRequest.Form.Get("refresh_token") == expectedRefreshToken &&
			httpRequest.Form.Get("client_secret") == expectedClientSecret {
			responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
			responseWriter.WriteHeader(http.StatusOK)
			_, _ = responseWriter.Write([]byte(
				`{"access_token": "newaccesstoken", "refresh_token": "new` + httpRequest.Form.Get("refresh_token") +
					`", "expires_in": 78901234}`))
		}
	})), &loginModuleCalledCount
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

//go:linkname contextWithParsedCookieParameters github.com/France-ioi/AlgoreaBackend/v2/app/api/auth.contextWithParsedCookieParameters
func contextWithParsedCookieParameters(ctx context.Context, cookieParameters *apiAuth.CookieParameters) context.Context
