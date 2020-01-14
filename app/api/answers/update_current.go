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
	// required: true
	// minLength: 1
	Answer string `json:"answer" validate:"set,min=1"`
	// required: true
	// minLength: 1
	State string `json:"state" validate:"set,min=1"`
}

// swagger:operation PUT /attempts/{attempt_id}/answers/current answers itemAnswerUpdateCurrent
// ---
// summary: Update current answer
// description: Update user's current answer. Used for auto-saving while working on a task.
//
//   * The authenticated user should have at least 'content' access to the `groups_attempts[attempt_id].item_id`
//
//   * `groups_attempts.group_id` should be the user's selfGroup or the user's team
//     [this extra check just ensures the consistency of data]
// parameters:
// - name: attempt_id
//   in: path
//   type: integer
//   required: true
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
	attemptID, err := service.ResolveURLQueryPathInt64Field(httpReq, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var requestData updateCurrentRequest
	formData := formdata.NewFormData(&requestData)
	err = formData.ParseJSONRequestData(httpReq)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)

	found, _, err := srv.Store.GroupAttempts().GetAttemptItemIDIfUserHasAccess(attemptID, user)
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		answersStore := store.Answers()
		service.MustNotBeError(answersStore.Where("answers.author_id = ?", user.GroupID).
			Where("answers.attempt_id = ?", attemptID).
			Where("answers.type = 'Current'").
			Delete().Error())

		return answersStore.RetryOnDuplicatePrimaryKeyError(func(store *database.DataStore) error {
			answerID := store.NewID()
			return store.Answers().InsertMap(map[string]interface{}{
				"id":         answerID,
				"author_id":  user.GroupID,
				"attempt_id": attemptID,
				"type":       "Current",
				"state":      requestData.State,
				"answer":     requestData.Answer,
				"created_at": database.Now(),
			})
		})
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.UpdateSuccess(nil)))
	return service.NoError
}
