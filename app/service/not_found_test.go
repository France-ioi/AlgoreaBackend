package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	assert := assertlib.New(t)
	req, _ := http.NewRequest("GET", "/dummy", http.NoBody)
	recorder := httptest.NewRecorder()

	NotFound(recorder, req)

	assert.Equal(`{"success":false,"message":"Not Found"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusNotFound, recorder.Code)
}
