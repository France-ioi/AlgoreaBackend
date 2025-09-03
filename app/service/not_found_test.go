package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/dummy", http.NoBody)
	recorder := httptest.NewRecorder()

	NotFound(recorder, req)

	assert.JSONEq(t, `{"success":false,"message":"Not Found"}`, recorder.Body.String())
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
