package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model
type updateCurrentRequest struct {
	// required:true
	AttemptID int64 `json:"attempt_id,string" validate:"required"`
	// required:true
	Answer string `json:"answer" validate:"required"`
	// required:true
	State string `json:"state" validate:"required"`
}

// swagger:operation PUT /answers/current answers itemAnswerUpdateCurrent
// ---
// summary: Update user’s current answer
// description: The service is used for auto-saving while working on a task.
//
//   * The authenticated user should have at least partial access to the groups_attempts[attempt_id].idItem
//
//   * groups_attempts.idGroup should be the user's selfGroup (if items.bHasAttempts=0) or the user's team (otherwise)
//   [this extra check just ensures the consistency of data]
// parameters:
// - name: current answer information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/updateCurrentRequest"
// responses:
//   "201":
//     "$ref": "#/responses/updatedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateCurrent(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	var requestData updateCurrentRequest

	formData := formdata.NewFormData(&requestData)
	err := formData.ParseJSONRequestData(httpReq)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	if err = user.Load(); err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	attemptID := requestData.AttemptID
	found, itemID, err := srv.Store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(attemptID, user)
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		userAnswerStore := store.UserAnswers()
		var currentAnswerID int64
		currentAnswerID, err = userAnswerStore.GetOrCreateCurrentAnswer(user.UserID, itemID, &attemptID)
		service.MustNotBeError(err)

		columnsToUpdate := map[string]interface{}{
			"sState":  requestData.State,
			"sAnswer": requestData.Answer,
		}
		service.MustNotBeError(userAnswerStore.ByID(currentAnswerID).UpdateColumn(columnsToUpdate).Error())

		service.MustNotBeError(store.UserItems().Where("idUser = ?", user.UserID).
			Where("idItem = ?", itemID).
			Where("idAttemptActive = ?", requestData.AttemptID).
			UpdateColumn(columnsToUpdate).Error())

		return nil
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.UpdateSuccess(nil)))
	return service.NoError
}
