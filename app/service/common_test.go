package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestBase_GetUser(t *testing.T) {
	db, mock := database.NewDBMock()
	mock.ExpectQuery("").WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"idGroupOwned"}).AddRow(int64(2)))
	middleware := auth.MockUserIDMiddleware(42)
	called := false
	ts := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		srv := &Base{Store: database.NewDataStore(db)}
		user := srv.GetUser(r)
		assert.Equal(t, int64(42), user.UserID)
		userOwnedGroupID, err := user.OwnedGroupID()
		assert.Equal(t, int64(2), userOwnedGroupID)
		assert.NoError(t, err)
	})))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	_, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)

	assert.True(t, called)
}

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
