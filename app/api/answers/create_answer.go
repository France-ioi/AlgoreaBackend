package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /items/{item_id}/attempts/{attempt_id}/answers answers answerCreate
//
//	---
//	summary: Create a "saved" answer
//	description: Creates a "saved" answer from a current snapshot.
//
//		- The authenticated user should have at least 'content' access to the `{item_id}`.
//
//		- `{as_team_id}` (if given) should be the user's team.
//
//		- There should be a row in the `results` table with `attempt_id` = `{attempt_id}`,
//			`participant_id` = the user's group (or `{as_team_id}` if given), `item_id` = `{item_id}`.
//
//		If any of the preconditions fails, the 'forbidden' error is returned.
//
//	parameters:
//		- name: attempt_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: answer information
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/answerData"
//
//	responses:
//
//		"201":
//			description: Created. The request has successfully saved the answer.
//			schema:
//				"$ref": "#/definitions/createdResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) answerCreate(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	return srv.saveAnswerWithType(responseWriter, httpRequest, false)
}

func (srv *Service) saveAnswerWithType(responseWriter http.ResponseWriter, httpRequest *http.Request, isCurrent bool) error {
	attemptID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var requestData answerData
	formData := formdata.NewFormData(&requestData)
	err = formData.ParseJSONRequestData(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	participantID := service.ParticipantIDFromContext(httpRequest.Context())
	store := srv.GetStore(httpRequest)

	found, err := store.Results().ByID(participantID, attemptID, itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	err = store.InTransaction(func(store *database.DataStore) error {
		answersStore := store.Answers()

		answerType := "Saved"
		if isCurrent {
			answerType = "Current"

			service.MustNotBeError(answersStore.Where("answers.author_id = ?", user.GroupID).
				Where("answers.attempt_id = ?", attemptID).
				Where("answers.participant_id = ?", participantID).
				Where("answers.item_id = ?", itemID).
				Where("answers.type = 'Current'").
				Delete().Error())
		}

		_, err = answersStore.CreateNewAnswer(user.GroupID, participantID, attemptID, itemID, answerType, requestData.Answer, &requestData.State)
		return err
	})
	service.MustNotBeError(err)

	var result render.Renderer
	if isCurrent {
		result = service.UpdateSuccess[*struct{}](nil)
	} else {
		result = service.CreationSuccess[*struct{}](nil)
	}

	service.MustNotBeError(render.Render(responseWriter, httpRequest, result))
	return nil
}
