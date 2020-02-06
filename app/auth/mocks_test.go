package auth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestMiddlewareMock(t *testing.T) {
	assert := assertlib.New(t)
	middleware := MockUserMiddleware(&database.User{GroupID: 42})
	ts := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		_, _ = w.Write([]byte(strconv.FormatInt(user.GroupID, 10)))
	})))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	response, err := http.DefaultClient.Do(request)
	assert.NoError(err)
	if err != nil {
		return
	}
	defer func() { _ = response.Body.Close() }()

	respBody, err := ioutil.ReadAll(response.Body)
	assert.NoError(err)
	assert.Equal("42", string(respBody))
}
