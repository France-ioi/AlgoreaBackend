package service

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestBase_ResolveWatchedGroupID(t *testing.T) {
	tests := []struct {
		name               string
		url                string
		wantWatchedGroupID int64
		wantOk             bool
		wantAPIError       APIError
	}{
		{name: "no watched_group_id", url: "/dummy", wantWatchedGroupID: 0, wantOk: false, wantAPIError: NoError},
		{
			name:               "invalid watched_group_id",
			url:                "/dummy?watched_group_id=abc",
			wantWatchedGroupID: 0,
			wantOk:             false,
			wantAPIError:       ErrInvalidRequest(errors.New("wrong value for watched_group_id (should be int64)")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			watchedGroupID, ok, apiError := (&Base{}).ResolveWatchedGroupID(req)

			assert.Equal(t, tt.wantWatchedGroupID, watchedGroupID)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantAPIError, apiError)
		})
	}
}

func TestBase_ResolveWatchedGroupID_DBError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	expectedError := errors.New("test")
	mock.ExpectQuery("").WillReturnError(expectedError)
	srv := &Base{store: database.NewDataStore(db)}
	patch := monkey.PatchInstanceMethod(reflect.TypeOf(srv), "GetUser",
		func(*Base, *http.Request) *database.User { return &database.User{GroupID: 567} })
	defer patch.Unpatch()

	req, _ := http.NewRequest("GET", "/dummy?watched_group_id=123", nil)
	watchedGroupID, ok, appErr := srv.ResolveWatchedGroupID(req)

	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, int64(0), watchedGroupID)
	assert.Equal(t, false, ok)
	assert.Equal(t, ErrUnexpected(expectedError), appErr)
}
