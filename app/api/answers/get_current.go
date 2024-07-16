package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /items/{item_id}/current-answer answers currentAnswerGet
//
//	---
//	summary: Get a current answer
//	description: Returns the latest auto-saved ('Current') answer for the given `{item_id}` and `{attempt_id}`.
//
//		* The user should have at least 'content' access rights to the `item_id` item.
//
//		* The user should be able to see answers related to his group's attempts so
//			the user should be a member of the `answers.participant_id` team or
//			`answers.participant_id` should be equal to the user's self group.
//
//		* `{as_team_id}` (if given) should be the user's team.
//
//
//		If any of the preconditions fails, the 'forbidden' error is returned.
//
//		If there is no current answer, the response is equal to `{"type":null}`.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: attempt_id
//			in: query
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			"$ref": "#/responses/itemAnswerGetResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getCurrentAnswer(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	attemptID, err := service.ResolveURLQueryGetInt64Field(httpReq, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	participantID := service.ParticipantIDFromContext(httpReq.Context())

	store := srv.GetStore(httpReq)
	user := srv.GetUser(httpReq)

	if !user.CanSeeAnswer(store, participantID, itemID) {
		return service.InsufficientAccessRightsError
	}

	answer, hasAnswer := store.Answers().GetCurrentAnswer(participantID, itemID, attemptID)
	if !hasAnswer {
		answer = map[string]interface{}{
			"type": nil,
		}
	}
	convertedResult := service.ConvertMapFromDBToJSON(answer)

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}
