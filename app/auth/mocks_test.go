package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

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
