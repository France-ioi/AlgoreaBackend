package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestUserMock(t *testing.T) {
	assert := assertlib.New(t)
	user := NewMockUser(1, 2, 3, 4)
	assert.Equal(int64(1), user.UserID)
	assert.Equal(int64(1), user.data.ID)
	assert.Equal(int64(2), user.data.SelfGroupID)
	assert.Equal(int64(3), user.data.OwnedGroupID)
	assert.Equal(int64(4), user.data.DefaultLanguageID)
}

func TestMiddlewareMock(t *testing.T) {
	assert := assertlib.New(t)
	middleware := MockUserIDMiddleware(42)
	ts := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		_, _ = w.Write([]byte(strconv.FormatInt(userID, 10)))
	})))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	response, err := http.DefaultClient.Do(request)
	assert.NoError(err)
	respBody, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Equal("42", string(respBody))
}
