package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	assertlib "github.com/stretchr/testify/assert"
)

func httpResponseForResponse(renderer render.Renderer) *httptest.ResponseRecorder {
	var fn AppHandler = func(respW http.ResponseWriter, req *http.Request) APIError {
		_ = render.Render(respW, req, renderer)
		return NoError
	}
	handler := http.HandlerFunc(fn.ServeHTTP)

	req, _ := http.NewRequest("GET", "/dummy", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestCreationSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		ItemID int64 `json:"ID"`
	}{42}

	recorder := httpResponseForResponse(CreationSuccess(data))
	assert.Equal(`{"success":true,"message":"success","data":{"ID":42}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusCreated, recorder.Code)
}
