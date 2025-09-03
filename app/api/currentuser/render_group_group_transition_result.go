package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// RenderGroupGroupTransitionResult renders database.GroupGroupTransitionResult as a response or returns an APIError.
func RenderGroupGroupTransitionResult(
	responseWriter http.ResponseWriter, httpRequest *http.Request, result database.GroupGroupTransitionResult,
	approvalsToRequest database.GroupApprovals, action userGroupRelationAction,
) error {
	isCreateAction := map[userGroupRelationAction]bool{
		createGroupJoinRequestAction:         true,
		joinGroupByCodeAction:                true,
		createAcceptedGroupJoinRequestAction: true,
		createGroupLeaveRequestAction:        true,
	}[action]
	switch result {
	case database.Cycle:
		return service.ErrUnprocessableEntity(errors.New("cycles in the group relations graph are not allowed"))
	case database.Invalid:
		if isCreateAction {
			return service.ErrUnprocessableEntity(errors.New("a conflicting relation exists"))
		}
		return service.ErrNotFound(errors.New("no such relation"))
	case database.Full:
		return service.ErrConflict(errors.New("the group is full"))
	case database.ApprovalsMissing:
		errorResponse := &service.ErrorResponse[map[string]interface{}]{
			Response: service.Response[map[string]interface{}]{
				HTTPStatusCode: http.StatusUnprocessableEntity,
				Success:        false,
				Message:        "Unprocessable Entity",
			},
			ErrorText: "Missing required approvals",
			Errors:    nil,
		}
		if approvalsToRequest != (database.GroupApprovals{}) {
			errorResponse.Data = map[string]interface{}{"missing_approvals": approvalsToRequest.ToArray()}
		}
		service.MustNotBeError(render.Render(responseWriter, httpRequest, errorResponse))
		return nil
	case database.Unchanged:
		statusCode := 200
		if isCreateAction {
			statusCode = 201
		}
		service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UnchangedSuccess(statusCode)))
	case database.Success:
		renderGroupGroupTransitionSuccess(isCreateAction,
			map[userGroupRelationAction]bool{
				leaveGroupAction: true,
			}[action], responseWriter, httpRequest)
	}
	return nil
}

func renderGroupGroupTransitionSuccess(isCreateAction, isDeleteAction bool, responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var successRenderer render.Renderer
	switch {
	case isCreateAction:
		successRenderer = service.CreationSuccess(map[string]bool{"changed": true})
	case isDeleteAction:
		successRenderer = service.DeletionSuccess(map[string]bool{"changed": true})
	default:
		successRenderer = service.UpdateSuccess(map[string]bool{"changed": true})
	}
	service.MustNotBeError(render.Render(responseWriter, httpRequest, successRenderer))
}
