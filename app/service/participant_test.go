package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
)

func TestGetParticipantIDFromRequest_NoAsTeamID(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	participantID, err := GetParticipantIDFromRequest(
		&http.Request{URL: &url.URL{}}, &database.User{GroupID: 123}, database.NewDataStore(db))
	assert.Equal(t, int64(123), participantID)
	assert.Nil(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetParticipantIDFromRequest_InvalidAsTeamID(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	participantID, err := GetParticipantIDFromRequest(
		&http.Request{URL: &url.URL{RawQuery: "as_team_id=abc"}}, &database.User{GroupID: 123}, database.NewDataStore(db))
	assert.Equal(t, int64(0), participantID)
	assert.Equal(t, ErrInvalidRequest(fmt.Errorf("wrong value for as_team_id (should be int64)")), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParticipantMiddleware(t *testing.T) {
	tests := []struct {
		name                     string
		asTeamID                 int64
		userID                   int64
		returnedError            error
		panicWith                interface{}
		expectedServiceWasCalled bool
		expectedStatusCode       int
		expectedBody             string
		logContains              string
	}{
		{
			name:                     "no as_team_id given",
			userID:                   890123,
			expectedServiceWasCalled: true,
			expectedStatusCode:       200,
			expectedBody:             "participant_id:890123",
		},
		{
			name:                     "as_team_id given",
			userID:                   890123,
			asTeamID:                 5678,
			expectedServiceWasCalled: true,
			expectedStatusCode:       200,
			expectedBody:             "participant_id:5678",
		},
		{
			name:                     "api error",
			userID:                   890123,
			asTeamID:                 5678,
			returnedError:            ErrForbidden(errors.New("some error")),
			expectedServiceWasCalled: false,
			expectedStatusCode:       403,
			expectedBody:             `{"success":false,"message":"Forbidden","error_text":"Some error"}`,
		},
		{
			name:                     "panic",
			userID:                   890123,
			asTeamID:                 5678,
			panicWith:                errors.New("some error"),
			expectedServiceWasCalled: false,
			expectedStatusCode:       500,
			expectedBody:             `{"success":false,"message":"Internal Server Error","error_text":"Unknown error"}`,
			logContains:              "unexpected error: some error",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := callThroughParticipantMiddleware(
				tt.userID, tt.asTeamID, tt.returnedError, tt.panicWith)
			defer func() {
				_ = result.resp.Body.Close()
			}()
			bodyBytes, _ := io.ReadAll(result.resp.Body)
			assert.Equal(t, tt.expectedStatusCode, result.resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", result.resp.Header.Get("Content-Type"))
			assert.True(t, result.middlewareWasCalled)
			assert.Equal(t, tt.expectedServiceWasCalled, result.serviceWasCalled)
			assert.Equal(t, result.actualUserID, tt.userID)
			assert.Contains(t, string(bodyBytes), tt.expectedBody)
			assert.NoError(t, result.mock.ExpectationsWereMet())
			assert.Contains(t, result.logsHook.GetAllLogs(), tt.logContains)
		})
	}
}

type callThroughParticipantMiddlewareResult struct {
	middlewareWasCalled bool
	serviceWasCalled    bool
	actualUserID        int64
	resp                *http.Response
	mock                sqlmock.Sqlmock
	logsHook            *loggingtest.Hook
}

func callThroughParticipantMiddleware(userID, asTeamID int64, returnedError error, panicWith interface{}) (
	result callThroughParticipantMiddlewareResult,
) {
	dbmock, mock := database.NewDBMock()
	defer func() { _ = dbmock.Close() }()
	userGuard := monkey.Patch(auth.UserFromContext, func(context.Context) *database.User {
		return &database.User{GroupID: userID}
	})
	defer userGuard.Unpatch()

	guard := monkey.Patch(GetParticipantIDFromRequest, func(_ *http.Request, user *database.User, _ *database.DataStore) (int64, error) {
		result.actualUserID = user.GroupID
		result.middlewareWasCalled = true
		if panicWith != nil {
			panic(panicWith)
		}
		if asTeamID != 0 {
			return asTeamID, returnedError
		}
		return result.actualUserID, returnedError
	})
	defer guard.Unpatch()

	hook, restoreFunc := logging.MockSharedLoggerHook()
	defer restoreFunc()

	// dummy server using the middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result.serviceWasCalled = true // has passed into the service
		body := "participant_id:" + strconv.FormatInt(ParticipantIDFromContext(r.Context()), 10)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	dataStore := database.NewDataStore(dbmock)
	participantMiddleware := ParticipantMiddleware(&Base{store: dataStore})
	mainSrv := httptest.NewServer(logging.NewStructuredLogger()(participantMiddleware(handler)))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest("GET", mainSrv.URL, http.NoBody)
	mainRequest.Header.Add("Authorization", "Bearer 1234567")
	client := &http.Client{}
	result.resp, _ = client.Do(mainRequest) //nolint:bodyclose // the body is closed in the test

	result.mock = mock
	result.logsHook = &loggingtest.Hook{Hook: hook}

	return result
}
