package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func TestMiddlewareMock(t *testing.T) {
	middleware := MockUserMiddleware(&database.User{GroupID: 42})
	testServer := httptest.NewServer(middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		assert.NotNil(t, SessionCookieAttributesFromContext(r.Context()))
		_, _ = w.Write([]byte(strconv.FormatInt(user.GroupID, 10)))
	})))
	defer testServer.Close()

	request, _ := http.NewRequest(http.MethodGet, testServer.URL, http.NoBody)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()

	respBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, "42", string(respBody))
}
