package service

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGetParticipantIDFromRequest_NoAsTeamID(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	participantID, apiError := GetParticipantIDFromRequest(
		&http.Request{URL: &url.URL{}}, &database.User{GroupID: 123}, database.NewDataStore(db))
	assert.Equal(t, int64(123), participantID)
	assert.Equal(t, NoError, apiError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetParticipantIDFromRequest_InvalidAsTeamID(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	participantID, apiError := GetParticipantIDFromRequest(
		&http.Request{URL: &url.URL{RawQuery: "as_team_id=abc"}}, &database.User{GroupID: 123}, database.NewDataStore(db))
	assert.Equal(t, int64(0), participantID)
	assert.Equal(t, ErrInvalidRequest(fmt.Errorf("wrong value for as_team_id (should be int64)")), apiError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParticipantMiddleware(t *testing.T) {
	tests := []struct {
		name                     string
		asTeamID                 int64
		userID                   int64
		apiError                 APIError
		expectedServiceWasCalled bool
		expectedStatusCode       int
		expectedBody             string
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
			apiError:                 ErrForbidden(errors.New("some error")),
			expectedServiceWasCalled: false,
			expectedStatusCode:       403,
			expectedBody:             `{"success":false,"message":"Forbidden","error_text":"Some error"}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			middlewareWasCalled, serviceWasCalled, actualUserID, resp, mock := callThroughParticipantMiddleware(tt.userID, tt.asTeamID, tt.apiError)
			defer func() { _ = resp.Body.Close() }()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
			assert.True(t, middlewareWasCalled)
			assert.Equal(t, tt.expectedServiceWasCalled, serviceWasCalled)
			assert.Equal(t, actualUserID, tt.userID)
			assert.Contains(t, string(bodyBytes), tt.expectedBody)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func callThroughParticipantMiddleware(userID, asTeamID int64, apiError APIError) (
	called, enteredService bool, actualUserID int64, resp *http.Response, mock sqlmock.Sqlmock) {
	dbmock, mock := database.NewDBMock()
	defer func() { _ = dbmock.Close() }()
	userGuard := monkey.Patch(auth.UserFromContext, func(context.Context) *database.User {
		return &database.User{GroupID: userID}
	})
	defer userGuard.Unpatch()

	guard := monkey.Patch(GetParticipantIDFromRequest, func(_ *http.Request, user *database.User, _ *database.DataStore) (int64, APIError) {
		actualUserID = user.GroupID
		called = true
		if asTeamID != 0 {
			return asTeamID, apiError
		}
		return actualUserID, apiError
	})
	defer guard.Unpatch()

	// dummy server using the middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enteredService = true // has passed into the service
		body := "participant_id:" + strconv.FormatInt(ParticipantIDFromContext(r.Context()), 10)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	dataStore := database.NewDataStore(dbmock)
	participantMiddleware := ParticipantMiddleware(&Base{store: dataStore})
	mainSrv := httptest.NewServer(participantMiddleware(handler))
	defer mainSrv.Close()

	// calling web server
	mainRequest, _ := http.NewRequest("GET", mainSrv.URL, nil)
	mainRequest.Header.Add("Authorization", "Bearer 1234567")
	client := &http.Client{}
	resp, _ = client.Do(mainRequest)

	return called, enteredService, actualUserID, resp, mock
}
