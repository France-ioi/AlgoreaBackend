package answers

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation PUT /items/{item_id}/attempts/{attempt_id}/answers/current answers currentAnswerUpdate
//
//	---
//	summary: Update current answer
//	description: Update participant's current answer. Used for auto-saving while working on a task.
//
//		* The authenticated user should have at least 'content' access to the `{item_id}`.
//
//		* `{as_team_id}` (if given) should be the user's team.
//
//		* There should be a row in the `results` table with `attempt_id` = `{attempt_id}`,
//			`participant_id` = the user's group (or `{as_team_id}` if given), `item_id` = `{item_id}`
//
//		If any of the preconditions fails, the 'forbidden' error is returned.
//	parameters:
//		- name: attempt_id
//			in: path
//			type: integer
//			required: true
//		- name: item_id
//			in: path
//			type: integer
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: current answer information
//			in: body
//			required: true
//			schema:
//				"$ref": "#/definitions/answerData"
//	responses:
//		"201":
//			"$ref": "#/responses/updatedResponse"
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
func (srv *Service) updateCurrentAnswer(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	return srv.saveAnswerWithType(rw, httpReq, true)
}
