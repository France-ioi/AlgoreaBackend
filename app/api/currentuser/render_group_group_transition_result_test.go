package currentuser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

func TestRenderGroupGroupTransitionResult(t *testing.T) {
	tests := []struct {
		name               string
		result             database.GroupGroupTransitionResult
		approvalsToRequest database.GroupApprovals
		actions            []userGroupRelationAction
		wantStatusCode     int
		wantResponseBody   string
	}{
		{
			name:           "cycle",
			result:         database.Cycle,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity",` +
				`"error_text":"Cycles in the group relations graph are not allowed"}`,
		},
		{
			name:           "full",
			result:         database.Full,
			wantStatusCode: http.StatusConflict,
			wantResponseBody: `{"success":false,"message":"Conflict",` +
				`"error_text":"The group is full"}`,
		},
		{
			name:             "invalid (not found)",
			result:           database.Invalid,
			actions:          []userGroupRelationAction{acceptInvitationAction, rejectInvitationAction, leaveGroupAction},
			wantStatusCode:   http.StatusNotFound,
			wantResponseBody: `{"success":false,"message":"Not Found","error_text":"No such relation"}`,
		},
		{
			name:           "invalid (unprocessable entity)",
			result:         database.Invalid,
			actions:        []userGroupRelationAction{createGroupJoinRequestAction, joinGroupByCodeAction},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity",` +
				`"error_text":"A conflicting relation exists"}`,
		},
		{
			name:             "unchanged (created)",
			result:           database.Unchanged,
			actions:          []userGroupRelationAction{createGroupJoinRequestAction, joinGroupByCodeAction},
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"success":true,"message":"unchanged","data":{"changed":false}}`,
		},
		{
			name:             "unchanged (ok)",
			result:           database.Unchanged,
			actions:          []userGroupRelationAction{acceptInvitationAction, rejectInvitationAction, leaveGroupAction},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"success":true,"message":"unchanged","data":{"changed":false}}`,
		},
		{
			name:             "success (updated)",
			result:           database.Success,
			actions:          []userGroupRelationAction{acceptInvitationAction, rejectInvitationAction},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"success":true,"message":"updated","data":{"changed":true}}`,
		},
		{
			name:             "success (created)",
			result:           database.Success,
			actions:          []userGroupRelationAction{createGroupJoinRequestAction, joinGroupByCodeAction},
			wantStatusCode:   http.StatusCreated,
			wantResponseBody: `{"success":true,"message":"created","data":{"changed":true}}`,
		},
		{
			name:             "success (deleted)",
			actions:          []userGroupRelationAction{leaveGroupAction},
			result:           database.Success,
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"success":true,"message":"deleted","data":{"changed":true}}`,
		},
		{
			name:             "approvals_missing",
			result:           database.ApprovalsMissing,
			wantStatusCode:   http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity","error_text":"Missing required approvals"}`,
		},
		{
			name:   "approvals_missing (with approvals listed)",
			result: database.ApprovalsMissing,
			approvalsToRequest: database.GroupApprovals{
				PersonalInfoViewApproval: true,
				LockMembershipApproval:   true,
				WatchApproval:            true,
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResponseBody: `{"success":false,"message":"Unprocessable Entity",` +
				`"data":{"missing_approvals":["personal_info_view","lock_membership","watch"]},` +
				`"error_text":"Missing required approvals"}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		if len(tt.actions) == 0 {
			tt.actions = []userGroupRelationAction{
				acceptInvitationAction, joinGroupByCodeAction, rejectInvitationAction,
				createGroupJoinRequestAction, leaveGroupAction,
			}
		}
		for _, action := range tt.actions {
			action := action
			t.Run(tt.name+": "+string(action), func(t *testing.T) {
				var fn service.AppHandler = func(respW http.ResponseWriter, req *http.Request) error {
					return RenderGroupGroupTransitionResult(respW, req, tt.result, tt.approvalsToRequest, action)
				}
				handler := http.HandlerFunc(fn.ServeHTTP)
				req, _ := http.NewRequest(http.MethodGet, "/dummy", http.NoBody)
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)

				assert.Equal(t, tt.wantStatusCode, recorder.Code)
				assert.Equal(t, tt.wantResponseBody, strings.TrimSpace(recorder.Body.String()))
			})
		}
	}
}
