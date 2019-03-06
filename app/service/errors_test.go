package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	assertlib "github.com/stretchr/testify/assert"
)

func responseForError(e APIError) *httptest.ResponseRecorder {
	return responseForHandler(func(writer http.ResponseWriter, r *http.Request) APIError {
		return e
	})
}

func responseForHandler(appHandler AppHandler) *httptest.ResponseRecorder {
	handler := http.HandlerFunc(appHandler.ServeHTTP)

	req, _ := http.NewRequest("GET", "/dummy", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
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
	recorder := responseForHandler(func(w http.ResponseWriter, r *http.Request) APIError {
		panic(errors.New("some error"))
	})
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Some error"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestRendersErrUnexpectedOnPanicWithSomeValue(t *testing.T) {
	assert := assertlib.New(t)
	expectedMessage := "some error"
	recorder := responseForHandler(func(w http.ResponseWriter, r *http.Request) APIError {
		panic(expectedMessage)
	})
	assert.Equal(`{"success":false,"message":"Internal Server Error","error_text":"Unknown error: `+expectedMessage+`"}`+"\n",
		recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}
