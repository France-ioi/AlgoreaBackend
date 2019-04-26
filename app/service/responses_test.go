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
	assert.Equal(`{"success":true,"message":"created","data":{"ID":42}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusCreated, recorder.Code)
}

func TestDeletionSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		Info string `json:"info"`
	}{"some info"}

	recorder := httpResponseForResponse(DeletionSuccess(data))
	assert.Equal(`{"success":true,"message":"deleted","data":{"info":"some info"}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestNotChangedSuccess(t *testing.T) {
	assert := assertlib.New(t)

	recorder := httpResponseForResponse(NotChangedSuccess())
	assert.Equal(`{"success":true,"message":"not changed"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusResetContent, recorder.Code)
}

func TestResponse_Render(t *testing.T) {
	response := &Response{HTTPStatusCode: http.StatusOK, Message: "", Success: true}
	recorder := httpResponseForResponse(response)
	assertlib.Equal(t, `{"success":true,"message":"success"}`+"\n", recorder.Body.String())
	assertlib.Equal(t, http.StatusOK, recorder.Code)
}
