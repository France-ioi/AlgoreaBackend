package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /attempts/{attempt_id}/answers answers itemAnswerSave
// ---
// summary: Save an answer
// description: Allows user to "save" a current snapshot of an answer manually.
//
//   * The authenticated user should have at least 'content' access to the `attempts[attempt_id].item_id`
//
//   * `attempts.group_id` should be the user or the user's team
//     [this extra check just ensures the consistency of data]
// parameters:
// - name: attempt_id
//   in: path
//   type: integer
//   required: true
// - name: answer information
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/answerData"
// responses:
//   "201":
//     description: Created. The request has successfully saved the answer.
//     schema:
//       "$ref": "#/definitions/createdResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) save(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	attemptID, err := service.ResolveURLQueryPathInt64Field(httpReq, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var requestData answerData
	formData := formdata.NewFormData(&requestData)
	err = formData.ParseJSONRequestData(httpReq)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)

	found, _, err := srv.Store.Attempts().GetAttemptItemIDIfUserHasAccess(attemptID, user)
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		answersStore := store.Answers()

		return answersStore.RetryOnDuplicatePrimaryKeyError(func(store *database.DataStore) error {
			answerID := store.NewID()
			return store.Answers().InsertMap(map[string]interface{}{
				"id":         answerID,
				"author_id":  user.GroupID,
				"attempt_id": attemptID,
				"type":       "Saved",
				"state":      requestData.State,
				"answer":     requestData.Answer,
				"created_at": database.Now(),
			})
		})
	})
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(rw, httpReq, service.CreationSuccess(nil)))
	return service.NoError
}
