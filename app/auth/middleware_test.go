package auth

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestUserIDMiddleware(t *testing.T) {
	tests := []struct {
		name                     string
		authHeaders              []string
		expectedAccessToken      string
		userIDReturnedByDB       int64
		dbError                  error
		expectedStatusCode       int
		expectedServiceWasCalled bool
		expectedBody             string
		expectedLogs             string
	}{
		{
			name:                     "valid access token",
			authHeaders:              []string{"Bearer 1234567"},
			expectedAccessToken:      "1234567",
			userIDReturnedByDB:       890123,
			expectedStatusCode:       200,
			expectedServiceWasCalled: true,
			expectedBody:             "user_id:890123",
		},
		{
			name:                     "missing access token",
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "No access token provided",
		},
		{
			name:                     "database error",
			authHeaders:              []string{"Bearer 123"},
			expectedAccessToken:      "123",
			dbError:                  errors.New("some error"),
			expectedStatusCode:       502,
			expectedServiceWasCalled: false,
			expectedBody:             "Can't validate the access token",
			expectedLogs:             `level=error msg="Can't validate an access token: some error"`,
		},
		{
			name:                     "expired token",
			authHeaders:              []string{"Bearer abcdefgh"},
			expectedAccessToken:      "abcdefgh",
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "Invalid access token",
		},
		{
			name:                     "spaces before the access token",
			authHeaders:              []string{"Bearer   1234567"},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "No access token provided",
		},
		{
			name:                     "spaces in access token",
			authHeaders:              []string{"Bearer 123 456 7"},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "No access token provided",
		},
		{
			name:                     "ignores other kinds of authorization headers",
			authHeaders:              []string{"Basic aladdin:opensesame"},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "No access token provided",
		},
		{
			name:                     "token is too long (should not query the DB)",
			authHeaders:              []string{"Bearer " + strings.Repeat("1", 256)},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "Invalid access token",
		},
		{
			name:                     "takes the first access token from headers",
			authHeaders:              []string{"Basic admin:password", "Bearer 1234567", "Bearer abcdefg"},
			expectedAccessToken:      "1234567",
			userIDReturnedByDB:       890123,
			expectedStatusCode:       200,
			expectedServiceWasCalled: true,
			expectedBody:             "user_id:890123",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert := assertlib.New(t)
			logHook, restoreFunc := logging.MockSharedLoggerHook()
			defer restoreFunc()

			serviceWasCalled, resp, mock := callAuthThroughMiddleware(tt.expectedAccessToken, tt.authHeaders, tt.userIDReturnedByDB, tt.dbError)
			defer func() { _ = resp.Body.Close() }()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			assert.Equal(tt.expectedStatusCode, resp.StatusCode)
			assert.Equal(tt.expectedServiceWasCalled, serviceWasCalled)
			assert.Contains(string(bodyBytes), tt.expectedBody)
			assert.Contains((&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs(), tt.expectedLogs)

			assert.NoError(mock.ExpectationsWereMet())
		})
	}
}

func callAuthThroughMiddleware(expectedSessionID string, authorizationHeaders []string,
	userID int64, dbError error) (bool, *http.Response, sqlmock.Sqlmock) {
	dbmock, mock := database.NewDBMock()
	defer func() { _ = dbmock.Close() }()
	if expectedSessionID != "" {
		expectation := mock.ExpectQuery("^" +
			regexp.QuoteMeta("SELECT idUser FROM `sessions`  WHERE (sAccessToken = ?) AND (sExpirationDate > NOW()) LIMIT 1") +
			"$").WithArgs(expectedSessionID)
		if dbError != nil {
			expectation.WillReturnError(dbError)
		} else {
			neededRows := mock.NewRows([]string{"idUser"})
			if userID != 0 {
				neededRows = neededRows.AddRow(userID)
			}
			expectation.WillReturnRows(neededRows)
		}
	}

	// dummy server using the middleware
	middleware := UserIDMiddleware(database.NewDataStore(dbmock).Sessions())
	enteredService := false // used to log if the service has been reached
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enteredService = true // has passed into the service
		userID := r.Context().Value(ctxUserID).(int64)
		body := "user_id:" + strconv.FormatInt(userID, 10)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	mainSrv := httptest.NewServer(middleware(handler))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest("GET", mainSrv.URL, nil)
	for _, header := range authorizationHeaders {
		mainRequest.Header.Add("Authorization", header)
	}
	client := &http.Client{}
	resp, _ := client.Do(mainRequest)

	return enteredService, resp, mock
}

func TestSetAuthorizationHeaderFromQueryMiddleware_PanicsInProduction(t *testing.T) {
	monkey.Patch(appenv.IsEnvProd, func() bool { return true })
	defer monkey.UnpatchAll()
	assertlib.Panics(t, func() {
		SetAuthorizationHeaderFromQueryMiddleware()
	})
}

func TestSetAuthorizationHeaderFromQueryMiddleware_SetsAuthorizationHeaderWhenAccessTokenPresent(t *testing.T) {
	middleware := SetAuthorizationHeaderFromQueryMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	assertlib.Equal(t, "Bearer 123",
		assertlib.HTTPBody(middleware(handler).ServeHTTP, "GET", "/", url.Values{"access_token": []string{"123"}}))
}

func TestSetAuthorizationHeaderFromQueryMiddleware_DoesNothingWhenAccessTokenAbsent(t *testing.T) {
	middleware := SetAuthorizationHeaderFromQueryMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	assertlib.Equal(t, "", assertlib.HTTPBody(middleware(handler).ServeHTTP, "GET", "/", nil))
}
