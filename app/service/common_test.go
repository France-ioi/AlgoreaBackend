package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestBase_GetUser(t *testing.T) {
	middleware := auth.MockUserMiddleware(&database.User{ID: 42, OwnedGroupID: ptrInt64(2)})
	called := false
	ts := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		srv := &Base{}
		user := srv.GetUser(r)
		assert.Equal(t, int64(42), user.ID)
		assert.Equal(t, ptrInt64(2), user.OwnedGroupID)
	})))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	_, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)

	assert.True(t, called)
}

func ptrInt64(i int64) *int64 { return &i }
