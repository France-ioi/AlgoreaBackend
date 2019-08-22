package currentuser

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// RenderGroupGroupTransitionResult renders database.GroupGroupTransitionResult as a response or returns an APIError
func RenderGroupGroupTransitionResult(w http.ResponseWriter, r *http.Request, result database.GroupGroupTransitionResult,
	action userGroupRelationAction) service.APIError {
	switch result {
	case database.Cycle:
		return service.ErrUnprocessableEntity(errors.New("cycles in the group relations graph are not allowed"))
	case database.Invalid:
		if action == createGroupRequestAction {
			return service.ErrUnprocessableEntity(errors.New("a conflicting relation exists"))
		}
		return service.ErrNotFound(errors.New("no such relation"))
	case database.Unchanged:
		statusCode := 200
		if action == createGroupRequestAction {
			statusCode = 201
		}
		service.MustNotBeError(render.Render(w, r, service.NotChangedSuccess(statusCode)))
	case database.Success:
		var successRenderer render.Renderer
		switch action {
		case leaveGroupAction:
			successRenderer = service.DeletionSuccess(nil)
		case createGroupRequestAction:
			successRenderer = service.CreationSuccess(nil)
		default:
			successRenderer = service.UpdateSuccess(nil)
		}
		service.MustNotBeError(render.Render(w, r, successRenderer))
	}
	return service.NoError
}
