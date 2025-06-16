package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/servicetest"
)

func responseForError(e error) *httptest.ResponseRecorder {
	return responseForHandler(func(http.ResponseWriter, *http.Request) error {
		return e
	})
}

func responseForHTTPHandler(handler http.Handler) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/dummy", http.NoBody)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
}

func responseForHandler(appHandler service.AppHandler) *httptest.ResponseRecorder {
	return responseForHTTPHandler(appHandler)
}

func TestNoErrorWithAPIError(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(&service.APIError{HTTPStatusCode: http.StatusConflict, EmbeddedError: nil})
	assert.Equal(`{"success":false,"message":"Conflict"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusConflict, recorder.Code)
}

func TestInvalidRequest(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrInvalidRequest(errors.New("sample invalid req")))
	assert.Equal(`{"success":false,"message":"Bad Request","error_text":"Sample invalid req"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestUnprocessableEntityRequest(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrUnprocessableEntity(errors.New("some error")))
	assert.Equal(`{"success":false,"message":"Unprocessable Entity","error_text":"Some error"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusUnprocessableEntity, recorder.Code)
}

func TestInvalidRequest_WithFormErrors(t *testing.T) {
	assert := assertlib.New(t)

	formErrors := make(formdata.FieldErrors)
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

func TestNotFound(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrNotFound(errors.New("some error")))
	assert.Equal(`{"success":false,"message":"Not Found","error_text":"Some error"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func TestRequestTimeout(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrRequestTimeout())
	assert.Equal(`{"success":false,"message":"Request Timeout"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusRequestTimeout, recorder.Code)
}

func TestConflict(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(service.ErrConflict(errors.New("conflict error")))
	assert.Equal(`{"success":false,"message":"Conflict","error_text":"Conflict error"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusConflict, recorder.Code)
}

func TestRendersErrUnexpectedOnPanicWithError(t *testing.T) {
	assert := assertlib.New(t)
	handler, hook, restoreFunc := servicetest.WithLoggingMiddleware(
		service.AppHandler(func(http.ResponseWriter, *http.Request) error {
			panic(errors.New("some error"))
		}))
	defer restoreFunc()

	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unknown error"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.GetAllLogs(), "unexpected error: some error")
}

func TestRendersRecoveredAPIErrorOnPanicWithAPIError(t *testing.T) {
	assert := assertlib.New(t)
	handler, hook, restoreFunc := servicetest.WithLoggingMiddleware(
		service.AppHandler(func(http.ResponseWriter, *http.Request) error {
			panic(service.InsufficientAccessRightsError)
		}))
	defer restoreFunc()

	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Forbidden","error_text":"Insufficient access rights"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusForbidden, recorder.Code)
	assert.NotContains(strings.ToLower(hook.GetAllLogs()), "error")
}

func TestRendersErrUnexpectedOnPanicWithSomeValue(t *testing.T) {
	assert := assertlib.New(t)
	expectedMessage := "some error"
	handler, hook, restoreFunc := servicetest.WithLoggingMiddleware(
		service.AppHandler(func(http.ResponseWriter, *http.Request) error {
			panic(expectedMessage)
		}))
	defer restoreFunc()

	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unknown error"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.GetAllLogs(), "unexpected error: some error")
}

func TestRendersErrRequestTimeoutOnPanicContextDeadlineExceeded(t *testing.T) {
	assert := assertlib.New(t)
	handler, _, restoreFunc := servicetest.WithLoggingMiddleware(
		service.AppHandler(func(http.ResponseWriter, *http.Request) error {
			panic(context.DeadlineExceeded)
		}))
	defer restoreFunc()

	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Request Timeout"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusRequestTimeout, recorder.Code)
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
