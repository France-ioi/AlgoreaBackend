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
		ItemID int64 `json:"id"`
	}{42}

	recorder := httpResponseForResponse(CreationSuccess(data))
	assert.Equal(`{"success":true,"message":"created","data":{"id":42}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusCreated, recorder.Code)
}

func TestUpdateSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		Info string `json:"info"`
	}{"some info"}

	recorder := httpResponseForResponse(UpdateSuccess(data))
	assert.Equal(`{"success":true,"message":"updated","data":{"info":"some info"}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
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

func TestUnchangedSuccess(t *testing.T) {
	assert := assertlib.New(t)

	recorder := httpResponseForResponse(UnchangedSuccess(http.StatusResetContent))
	assert.Equal(`{"success":true,"message":"unchanged","data":{"changed":false}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusResetContent, recorder.Code)
}

func TestResponse_Render(t *testing.T) {
	response := &Response{HTTPStatusCode: http.StatusOK, Message: "", Success: true}
	recorder := httpResponseForResponse(response)
	assertlib.Equal(t, `{"success":true,"message":"success"}`+"\n", recorder.Body.String())
	assertlib.Equal(t, http.StatusOK, recorder.Code)
}
