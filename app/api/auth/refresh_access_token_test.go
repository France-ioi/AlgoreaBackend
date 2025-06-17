package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/servicetest"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestService_refreshAccessToken_NotAllowRefreshTokenRaces(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	expectedClientID := "1234"
	expectedClientSecret := "secret"
	loginModuleStubServer := createLoginModuleStubServer(expectedClientID, expectedClientSecret)
	done := make(chan bool)
	doRequest := func(timeout bool) {
		var cancelFunc context.CancelFunc
		if timeout {
			var patchGuard *monkey.PatchGuard
			patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&sessionIDsInProgressMap{}), "WithLock",
				func(m *sessionIDsInProgressMap, sessionID int64, r *http.Request, f func() error) error {
					cancelFunc()
					ctx := r.Context()
					(*valueCtxInterface)(unsafe.Pointer(&ctx)).p.timerCtx.err = context.DeadlineExceeded //nolint:gosec // imitate a timeout
					patchGuard.Unpatch()
					defer patchGuard.Restore()
					return m.WithLock(sessionID, r, f)
				})
			defer patchGuard.Unpatch()
		}
		response, mock, logs, err := servicetest.GetResponseForRouteWithMockedDBAndUser(
			"POST", "/auth/token", "", &database.User{GroupID: 2},
			func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^" +
					regexp.QuoteMeta(
						"SELECT "+
							"token, "+
							"TIMESTAMPDIFF(SECOND, NOW(), expires_at) AS seconds_until_expiry, "+
							"issued_at > (NOW() - INTERVAL 5 MINUTE) AS too_new_to_refresh "+
							"FROM `access_tokens`  WHERE (session_id = ?) ORDER BY expires_at DESC LIMIT 1") + "$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(mock.NewRows([]string{"token", "seconds_until_expiry", "too_new_to_refresh"}).
						AddRow("accesstoken", 600, false))

				if !timeout {
					mock.ExpectQuery("^" +
						regexp.QuoteMeta("SELECT refresh_token FROM `sessions` WHERE (session_id = ?) LIMIT 1") + "$").
						WithArgs(sqlmock.AnyArg()).
						WillReturnRows(mock.NewRows([]string{"refresh_token"}).AddRow("firstrefreshtoken"))
					mock.ExpectBegin()
					mock.ExpectExec("^"+regexp.QuoteMeta(
						"INSERT INTO `access_tokens` (`expires_at`, `issued_at`, `session_id`, `token`) "+
							"VALUES (NOW() + INTERVAL ? SECOND, NOW(), ?, ?)")+"$").
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "newaccesstoken").
						WillReturnResult(sqlmock.NewResult(123, 1))
					mock.ExpectExec("^"+regexp.QuoteMeta("UPDATE `sessions` SET `refresh_token` = ? WHERE (session_id = ?)")+"$").
						WithArgs("newfirstrefreshtoken", sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(-1, 1))
					mock.ExpectCommit()
					mock.ExpectExec("^DELETE access_tokens .+$").
						WithArgs(sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(-1, 1))
				}
			},
			func(mux *chi.Mux, baseService *service.Base) {
				srv := &Service{Base: baseService}
				srv.AuthConfig = viper.New()
				srv.AuthConfig.Set("loginModuleURL", loginModuleStubServer.URL)
				srv.AuthConfig.Set("clientID", expectedClientID)
				srv.AuthConfig.Set("clientSecret", expectedClientSecret)
				var router chi.Router = mux
				if timeout {
					router = router.With(func(next http.Handler) http.Handler {
						fn := func(w http.ResponseWriter, r *http.Request) {
							var ctx context.Context
							ctx, cancelFunc = context.WithDeadline(r.Context(), time.Now().Add(1*time.Hour))
							defer cancelFunc()

							r = r.WithContext(ctx)
							next.ServeHTTP(w, r)
						}
						return http.HandlerFunc(fn)
					})
				}
				router.
					With(middleware.WithValue(parsedRequestData, map[string]interface{}{})).
					Post("/auth/token", service.AppHandler(srv.refreshAccessToken).ServeHTTP)
			})
		assert.NoError(t, err)
		if err == nil {
			defer func() { _ = response.Body.Close() }()
		}
		if timeout {
			assert.Equal(t, 408, response.StatusCode)
			assert.Contains(t, logs, "The request is canceled: context deadline exceeded")
		} else {
			assert.Equal(t, 201, response.StatusCode)
			body, _ := io.ReadAll(response.Body)
			assert.Equal(t,
				`{"success":true,"message":"created","data":{"access_token":"newaccesstoken","expires_in":78901234}}`+"\n",
				string(body))
		}
		assert.NoError(t, mock.ExpectationsWereMet())
		done <- true
	}

	// check that the service waits while the session is locked
	mutexChannel := make(chan bool, 1)
	(*sync.Map)(&sessionIDsInProgress).Store(auth.MockCtxSessionID, mutexChannel) // lock the session
	mutexChannel <- true
	go doRequest(false)
	mutexChannel <- true // wait until refreshAccessToken() reads from the channel (meaning the service is inside the for loop)
	close(mutexChannel)
	(*sync.Map)(&sessionIDsInProgress).Delete(auth.MockCtxSessionID) // here the service gets unlocked
	<-done                                                           // wait until the service finishes

	// check that the service timeouts if the session is locked for too long
	mutexChannel = make(chan bool, 1)
	(*sync.Map)(&sessionIDsInProgress).Store(auth.MockCtxSessionID, mutexChannel) // lock the session
	// Remove the mutex once we're finished, otherwise it makes further tests block if they use the same session ID.
	defer (*sync.Map)(&sessionIDsInProgress).Delete(auth.MockCtxSessionID)
	mutexChannel <- true
	go doRequest(true)
	<-done // wait until the service finishes
}

func createLoginModuleStubServer(expectedClientID, expectedClientSecret string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path == "/oauth/token" &&
			r.Form.Get("grant_type") == "refresh_token" &&
			r.Form.Get("client_id") == expectedClientID &&
			r.Form.Get("client_secret") == expectedClientSecret {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(
				`{"access_token": "newaccesstoken", "refresh_token": "new` + r.Form.Get("refresh_token") +
					`", "expires_in": 78901234}`))
		}
	}))
}

type cancelCtx struct {
	context.Context
	_   sync.Mutex
	_   atomic.Value
	_   map[interface{}]struct{}
	err error
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
