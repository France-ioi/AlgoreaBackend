package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestBase_SetQueryLimit(t *testing.T) {
	testCases := []struct {
		desc          string
		queryString   string
		expectedValue int64
	}{
		{
			desc:          "with limit",
			queryString:   "?limit=10",
			expectedValue: 10,
		},
		{
			desc:          "without limit",
			queryString:   "/",
			expectedValue: 500,
		},
		{
			desc:          "wrong limit",
			queryString:   "/?limit=any",
			expectedValue: 500,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			r := chi.NewRouter()
			called := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				db, mock := database.NewDBMock()
				defer func() { _ = db.Close() }()

				mock.ExpectQuery(regexp.QuoteMeta(
					fmt.Sprintf("SELECT * FROM `users` LIMIT %d", testCase.expectedValue))).
					WillReturnRows(mock.NewRows([]string{"ID"}))

				var result []interface{}
				err := SetQueryLimit(r, db.Table("users")).Scan(&result).Error()
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
			r.Get("/", handler)

			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, nil)
			_, _ = http.DefaultClient.Do(request)
			ts.Close()
			assert.True(t, called, "The handler was not called")
		})
	}
}
