package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestRenderGroupGroupTransitionResult(t *testing.T) {
	tests := []struct {
		name                              string
		result                            database.GroupGroupTransitionResult
		treatInvalidAsUnprocessableEntity bool
		treatSuccessAdDeleted             bool
		wantStatusCode                    int
		wantResponseBody                  string
	}{
		{
			name:           "cycle",
			result:         database.Cycle,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity",` +
				`"error_text":"Cycles in the group relations graph are not allowed"}`,
		},
		{
			name:             "invalid (not found)",
			result:           database.Invalid,
			wantStatusCode:   http.StatusNotFound,
			wantResponseBody: `{"success":false,"message":"Not Found","error_text":"No such relation"}`,
		},
		{
			name:                              "invalid (unprocessable entity)",
			result:                            database.Invalid,
			treatInvalidAsUnprocessableEntity: true,
			wantStatusCode:                    http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity",` +
				`"error_text":"A conflicting relation exists"}`,
		},
		{
			name:             "unchanged (created)",
			result:           database.Unchanged,
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"success":true,"message":"not changed"}`,
		},
		{
			name:                  "unchanged (ok)",
			result:                database.Unchanged,
			treatSuccessAdDeleted: true,
			wantStatusCode:        http.StatusOK,
			wantResponseBody:      `{"success":true,"message":"not changed"}`,
		},
		{
			name:             "success (created)",
			result:           database.Success,
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"success":true,"message":"created"}`,
		},
		{
			name:                  "success (deleted)",
			treatSuccessAdDeleted: true,
			result:                database.Success,
			wantStatusCode:        http.StatusOK,
			wantResponseBody:      `{"success":true,"message":"deleted"}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var fn AppHandler = func(respW http.ResponseWriter, req *http.Request) APIError {
				return RenderGroupGroupTransitionResult(respW, req, tt.result,
					tt.treatInvalidAsUnprocessableEntity, tt.treatSuccessAdDeleted)
			}
			handler := http.HandlerFunc(fn.ServeHTTP)
			req, _ := http.NewRequest("GET", "/dummy", nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantStatusCode, recorder.Code)
			assert.Equal(t, tt.wantResponseBody, strings.TrimSpace(recorder.Body.String()))
		})
	}
}
