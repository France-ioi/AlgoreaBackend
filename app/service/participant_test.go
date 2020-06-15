package service

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

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
