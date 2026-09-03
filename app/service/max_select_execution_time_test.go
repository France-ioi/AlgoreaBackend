package service_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/servicetest"
)

func TestRendersErrRequestTimeoutOnPanicQueryTimeout(t *testing.T) {
	tassert := assert.New(t)
	handler, _ := servicetest.WithLoggingMiddleware(
		service.AppHandler(func(http.ResponseWriter, *http.Request) error {
			panic(&mysql.MySQLError{
				Number:  uint16(mysqldb.QueryTimeoutError),
				Message: "Query execution was interrupted, maximum statement execution time exceeded",
			})
		}))

	recorder := responseForHTTPHandler(handler)
	tassert.JSONEq(`{"success":false,"message":"Request Timeout"}`, recorder.Body.String())
	tassert.Equal(http.StatusRequestTimeout, recorder.Code)
}

func TestMaxSelectExecutionTimeMiddleware(t *testing.T) {
	var got time.Duration
	handler := service.MaxSelectExecutionTime(1500 * time.Millisecond)(http.HandlerFunc(
		func(_ http.ResponseWriter, r *http.Request) {
			got = database.MaxSelectExecutionTimeFromContext(r.Context())
		}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assert.Equal(t, 1500*time.Millisecond, got)
}
