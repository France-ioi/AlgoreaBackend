//go:build !prod

// Package servicetest provides utilities to test services.
package servicetest

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// GetResponseForRouteWithMockedDBAndUser executes a route for unit tests
// auth.UserIDFromContext is stubbed to return the given userID.
// The test should provide functions that prepare the router and the sql mock.
func GetResponseForRouteWithMockedDBAndUser(
	method, path, requestBody string, user *database.User,
	setMockExpectationsFunc func(sqlmock.Sqlmock),
	setRouterFunc func(router *chi.Mux, baseService *service.Base),
) (*http.Response, sqlmock.Sqlmock, string, error) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	setMockExpectationsFunc(mock)

	base := service.Base{}
	base.SetGlobalStore(database.NewDataStore(db))
	router := chi.NewRouter()
	logger, logHook := logging.NewMockLogger()
	router.Use(logging.ContextWithLoggerMiddleware(logger))
	router.Use(auth.MockUserMiddleware(user))
	router.Use(middleware.RequestLogger(&logging.StructuredLogger{}))
	setRouterFunc(router, &base)

	ts := httptest.NewServer(router)
	defer ts.Close()

	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(requestBody))
	var response *http.Response
	if err == nil {
		response, err = http.DefaultClient.Do(request)
	}

	return response, mock, (&loggingtest.Hook{Hook: logHook}).GetAllLogs(), err
}

// WithLoggingMiddleware wraps the given handler in NullLogger with hook.
func WithLoggingMiddleware(appHandler http.Handler) (http.Handler, *loggingtest.Hook) {
	logger, hook := logging.NewMockLogger()
	loggingMiddleware := middleware.RequestLogger(&logging.StructuredLogger{})
	return logging.ContextWithLoggerMiddleware(logger)(loggingMiddleware(appHandler)), &loggingtest.Hook{Hook: hook}
}
