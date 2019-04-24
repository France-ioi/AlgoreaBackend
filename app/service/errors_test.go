package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/servicetest"
)

func responseForError(e service.APIError) *httptest.ResponseRecorder {
	return responseForHandler(func(http.ResponseWriter, *http.Request) service.APIError {
		return e
	})
}

func responseForHTTPHandler(handler http.Handler) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/dummy", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
}

func responseForHandler(appHandler service.AppHandler) *httptest.ResponseRecorder {
	return responseForHTTPHandler(appHandler)
}

func TestNoErrorWithAPIError(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.APIError{HTTPStatusCode: http.StatusConflict, Error: nil})
	assert.Equal(`{"success":false,"message":"Conflict"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusConflict, recorder.Code)
}

func TestInvalidRequest(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrInvalidRequest(errors.New("sample invalid req")))
	assert.Equal(`{"success":false,"message":"Bad Request","error_text":"Sample invalid req"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestInvalidRequest_WithFormErrors(t *testing.T) {
	assert := assertlib.New(t)

	formErrors := make(service.FieldErrors)
	formErrors["name"] = []string{"is required"}
	formErrors["phone"] = []string{"is required", "must be a phone number"}

	recorder := responseForError(service.ErrInvalidRequest(formErrors))
	assert.JSONEq(`{
			"success":false,
			"message":"Bad Request",
			"error_text":"Invalid input data",
			"errors": {
				"name": ["is required"],
				"phone": [
					"is required",
					"must be a phone number"
				]
			}
	}`, recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestForbidden(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrForbidden(errors.New("sample forbidden resp")))
	assert.Equal(`{"success":false,"message":"Forbidden","error_text":"Sample forbidden resp"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusForbidden, recorder.Code)
}

func TestUnexpected(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrUnexpected(errors.New("unexp err")))
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unexp err"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestRendersErrUnexpectedOnPanicWithError(t *testing.T) {
	assert := assertlib.New(t)
	handler, hook := servicetest.WithLoggingMiddleware(service.AppHandler(func(http.ResponseWriter, *http.Request) service.APIError {
		panic(errors.New("some error"))
	}))
	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Some error"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.GetAllLogs(), "unexpected error: some error")
}

func TestRendersErrUnexpectedOnPanicWithSomeValue(t *testing.T) {
	assert := assertlib.New(t)
	expectedMessage := "some error"
	handler, hook := servicetest.WithLoggingMiddleware(service.AppHandler(func(http.ResponseWriter, *http.Request) service.APIError {
		panic(expectedMessage)
	}))
	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unknown error: `+expectedMessage+`"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.GetAllLogs(), "unexpected error: unknown error: some error")
}

func TestMustNotBeError_PanicsOnError(t *testing.T) {
	expectedError := errors.New("some error")
	assertlib.PanicsWithValue(t, expectedError, func() {
		service.MustNotBeError(expectedError)
	})
}

func TestMustNotBeError_NotPanicsIfNoError(t *testing.T) {
	assertlib.NotPanics(t, func() {
		service.MustNotBeError(nil)
	})
}
