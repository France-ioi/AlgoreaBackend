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

func TestQueryLimiter_Apply(t *testing.T) {
	testCases := []struct {
		desc          string
		queryString   string
		queryLimiter  *QueryLimiter
		expectedValue int64
	}{
		{
			desc:          "with limit",
			queryString:   "?limit=10",
			queryLimiter:  NewQueryLimiter(),
			expectedValue: 10,
		},
		{
			desc:          "without limit",
			queryString:   "/",
			queryLimiter:  NewQueryLimiter(),
			expectedValue: 500,
		},
		{
			desc:          "wrong limit",
			queryString:   "/?limit=any",
			queryLimiter:  NewQueryLimiter(),
			expectedValue: 500,
		},
		{
			desc:          "without limit, but with a custom default value limit",
			queryString:   "/",
			queryLimiter:  NewQueryLimiter().SetDefaultLimit(400),
			expectedValue: 400,
		},
		{
			desc:          "wrong limit and a custom default limit",
			queryString:   "/?limit=any",
			queryLimiter:  NewQueryLimiter().SetDefaultLimit(10),
			expectedValue: 10,
		},
		{
			desc:          "limit is less than the maximum allowed limit",
			queryString:   "?limit=10",
			queryLimiter:  NewQueryLimiter().SetMaxAllowedLimit(11),
			expectedValue: 10,
		},
		{
			desc:          "limit exceeds the maximum allowed limit",
			queryString:   "?limit=10",
			queryLimiter:  NewQueryLimiter().SetMaxAllowedLimit(9),
			expectedValue: 9,
		},
		{
			desc:          "default limit exceeds the maximum allowed limit",
			queryString:   "/",
			queryLimiter:  NewQueryLimiter().SetDefaultLimit(500).SetMaxAllowedLimit(400),
			expectedValue: 400,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.desc, func(t *testing.T) {
			r := chi.NewRouter()
			called := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				db, mock := database.NewDBMock()
				defer func() { _ = db.Close() }()

				mock.ExpectQuery(regexp.QuoteMeta(
					fmt.Sprintf("SELECT * FROM `users` LIMIT %d", testCase.expectedValue))).
					WillReturnRows(mock.NewRows([]string{"id"}))

				var result []interface{}
				err := testCase.queryLimiter.Apply(r, db.Table("users")).Scan(&result).Error()
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
			r.Get("/", handler)

			ts := httptest.NewServer(r)
			request, _ := http.NewRequest("GET", ts.URL+testCase.queryString, http.NoBody)
			response, err := http.DefaultClient.Do(request)
			if err == nil {
				_ = response.Body.Close()
			}
			ts.Close()
			assert.True(t, called, "The handler was not called")
		})
	}
}
