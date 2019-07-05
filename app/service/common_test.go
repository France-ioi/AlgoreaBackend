package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
