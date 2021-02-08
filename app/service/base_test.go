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
	middleware := auth.MockUserMiddleware(&database.User{GroupID: 42})
	called := false
	ts := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		srv := &Base{}
		user := srv.GetUser(r)
		assert.Equal(t, int64(42), user.GroupID)
	})))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	response, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	if err == nil {
		_ = response.Body.Close()
	}

	assert.True(t, called)
}
