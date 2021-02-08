package auth

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/loggingtest"
)

func TestUserMiddleware(t *testing.T) {
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
			expectedBody:             "user_id:890123\nBearer:1234567",
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
			expectedStatusCode:       500,
			expectedServiceWasCalled: false,
			expectedBody:             `{"success":false,"message":"Internal server error","error_text":"Can't validate the access token"}` + "\n",
			expectedLogs:             `level=error msg="Can't validate an access token: some error"`,
		},
		{
			name:                     "expired token",
			authHeaders:              []string{"Bearer abcdefgh"},
			expectedAccessToken:      "abcdefgh",
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             `{"success":false,"message":"Unauthorized","error_text":"Invalid access token"}` + "\n",
		},
		{
			name:                     "spaces before the access token",
			authHeaders:              []string{"Bearer   1234567"},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             `{"success":false,"message":"Unauthorized","error_text":"No access token provided"}` + "\n",
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
			authHeaders:              []string{"Bearer " + strings.Repeat("1", 2001)},
			expectedStatusCode:       401,
			expectedServiceWasCalled: false,
			expectedBody:             "Invalid access token",
		},
		{
			name:                     "works fine with long tokens",
			authHeaders:              []string{"Bearer " + strings.Repeat("1", 2000)},
			expectedStatusCode:       200,
			expectedAccessToken:      strings.Repeat("1", 2000),
			userIDReturnedByDB:       78234,
			expectedServiceWasCalled: true,
			expectedBody:             "user_id:78234\nBearer:" + strings.Repeat("1", 2000),
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
		{
			name:                     "sets user attributes",
			authHeaders:              []string{"Bearer 1234567"},
			expectedAccessToken:      "1234567",
			userIDReturnedByDB:       890123,
			expectedStatusCode:       200,
			expectedServiceWasCalled: true,
			expectedBody: `User:{"GroupID":890123,"Login":"login","LoginID":12345,"DefaultLanguage":"fr",` +
				`"IsAdmin":true,"IsTempUser":true,"AccessGroupID":23456,"AllowSubgroups":true,"NotificationsReadAt":"2019-05-30T11:00:00Z"}`,
		},
		{
			name:                     "sets cookie attributes",
			authHeaders:              []string{"Bearer 1234567"},
			expectedAccessToken:      "1234567",
			userIDReturnedByDB:       890123,
			expectedStatusCode:       200,
			expectedServiceWasCalled: true,
			expectedBody: `CookieAttributes:{"UseCookie":true,"Secure":true,"SameSite":true,"Domain":"somedomain.org",` +
				`"Path":"/api/"}`,
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
			assert.Equal("application/json; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.Equal(tt.expectedServiceWasCalled, serviceWasCalled)
			assert.Contains(string(bodyBytes), tt.expectedBody)
			logs := (&loggingtest.Hook{Hook: logHook}).GetAllStructuredLogs()
			if tt.expectedLogs == "" {
				assert.Empty(logs)
			} else {
				assert.Contains(logs, tt.expectedLogs)
			}
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
			regexp.QuoteMeta(
				"SELECT users.login, users.login_id, users.is_admin, users.group_id, users.access_group_id, "+
					"users.temp_user, users.allow_subgroups, users.notifications_read_at, users.default_language, "+
					"sessions.use_cookie, sessions.cookie_secure, sessions.cookie_same_site, sessions.cookie_domain, sessions.cookie_path "+
					"FROM `sessions` "+
					"JOIN users ON users.group_id = sessions.user_id "+
					"WHERE (access_token = ?) AND (expires_at > NOW()) LIMIT 1") +
			"$").WithArgs(expectedSessionID)
		if dbError != nil {
			expectation.WillReturnError(dbError)
		} else {
			neededRows := mock.NewRows([]string{"group_id", "login", "login_id", "is_admin", "access_group_id", "temp_user",
				"allow_subgroups", "notifications_read_at", "default_language", "use_cookie", "cookie_secure", "cookie_same_site",
				"cookie_domain", "cookie_path"})
			if userID != 0 {
				neededRows = neededRows.AddRow(userID, "login", "12345", int64(1), int64(23456), int64(1), int64(1),
					[]byte("2019-05-30 11:00:00"), "fr", int64(1), int64(1), int64(1), "somedomain.org", "/api/")
			}
			expectation.WillReturnRows(neededRows)
		}
	}

	// dummy server using the middleware
	middleware := UserMiddleware(database.NewDataStore(dbmock).Sessions())
	enteredService := false // used to log if the service has been reached
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enteredService = true // has passed into the service
		user := r.Context().Value(ctxUser).(*database.User)
		cookieAttributes, _ := json.Marshal(r.Context().Value(ctxSessionCookieAttributes))
		userAttributes, _ := json.Marshal(r.Context().Value(ctxUser))
		body := "user_id:" + strconv.FormatInt(user.GroupID, 10) + "\nBearer:" + r.Context().Value(ctxBearer).(string) +
			"\nCookieAttributes:" + string(cookieAttributes) + "\nUser:" + string(userAttributes)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
