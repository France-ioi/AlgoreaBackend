package service

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// RenderGroupGroupTransitionResult renders database.GroupGroupTransitionResult as a response or returns an APIError
func RenderGroupGroupTransitionResult(w http.ResponseWriter, r *http.Request, result database.GroupGroupTransitionResult,
	treatInvalidAsUnprocessableEntity, treatSuccessAsDeleted bool) APIError {
	switch result {
	case database.Cycle:
		return ErrUnprocessableEntity(errors.New("cycles in the group relations graph are not allowed"))
	case database.Invalid:
		if treatInvalidAsUnprocessableEntity {
			return ErrUnprocessableEntity(errors.New("a conflicting relation exists"))
		}
		return ErrNotFound(errors.New("no such relation"))
	case database.Unchanged:
		if treatSuccessAsDeleted {
			MustNotBeError(render.Render(w, r, NotChangedSuccess(http.StatusOK)))
		} else {
			MustNotBeError(render.Render(w, r, NotChangedSuccess(http.StatusCreated)))
		}
	case database.Success:
		if treatSuccessAsDeleted {
			MustNotBeError(render.Render(w, r, DeletionSuccess(nil)))
		} else {
			MustNotBeError(render.Render(w, r, CreationSuccess(nil)))
		}
	}
	return NoError
}
