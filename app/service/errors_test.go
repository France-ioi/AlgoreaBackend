package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus/hooks/test"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

func responseForError(e APIError) *httptest.ResponseRecorder {
	return responseForHandler(func(_ http.ResponseWriter, _ *http.Request) APIError {
		return e
	})
}

func responseForHTTPHandler(handler http.Handler) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/dummy", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
}

func responseForHandler(appHandler AppHandler) *httptest.ResponseRecorder {
	return responseForHTTPHandler(appHandler)
}

func withLoggingMiddleware(appHandler AppHandler) (http.Handler, *test.Hook) {
	logger, hook := test.NewNullLogger()
	middleware := middleware.RequestLogger(&logging.StructuredLogger{Logger: logger})
	return middleware(appHandler), hook
}

func TestNoErrorWithAPIError(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(APIError{http.StatusConflict, nil})
	assert.Equal(`{"success":false,"message":"Conflict"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusConflict, recorder.Code)
}

func TestInvalidRequest(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(ErrInvalidRequest(errors.New("sample invalid req")))
	assert.Equal(`{"success":false,"message":"Bad Request","error_text":"Sample invalid req"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestInvalidRequest_WithFormErrors(t *testing.T) {
	assert := assertlib.New(t)

	formErrors := make(FieldErrors)
	formErrors["name"] = []string{"is required"}
	formErrors["phone"] = []string{"is required", "must be a phone number"}

	recorder := responseForError(ErrInvalidRequest(formErrors))
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
	recorder := responseForError(ErrForbidden(errors.New("sample forbidden resp")))
	assert.Equal(`{"success":false,"message":"Forbidden","error_text":"Sample forbidden resp"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusForbidden, recorder.Code)
}

func TestUnexpected(t *testing.T) {
	assert := assertlib.New(t)
	recorder := responseForError(ErrUnexpected(errors.New("unexp err")))
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unexp err"}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestRendersErrUnexpectedOnPanicWithError(t *testing.T) {
	assert := assertlib.New(t)
	handler, hook := withLoggingMiddleware(func(_ http.ResponseWriter, _ *http.Request) APIError {
		panic(errors.New("some error"))
	})
	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Some error"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.Entries[1].Message, "unexpected error: some error")
}

func TestRendersErrUnexpectedOnPanicWithSomeValue(t *testing.T) {
	assert := assertlib.New(t)
	expectedMessage := "some error"
	handler, hook := withLoggingMiddleware(func(_ http.ResponseWriter, _ *http.Request) APIError {
		panic(expectedMessage)
	})
	recorder := responseForHTTPHandler(handler)
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unknown error: `+expectedMessage+`"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
	assert.Contains(hook.Entries[1].Message, "unexpected error: unknown error: some error")
}

func TestMustNotBeError_PanicsOnError(t *testing.T) {
	expectedError := errors.New("some error")
	assertlib.PanicsWithValue(t, expectedError, func() {
		MustNotBeError(expectedError)
	})
}

func TestMustNotBeError_NotPanicsIfNoError(t *testing.T) {
	assertlib.NotPanics(t, func() {
		MustNotBeError(nil)
	})
}
