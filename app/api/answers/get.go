package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /answers/{answer_id} answers itemAnswerGet
// ---
// summary: Get an answer
// description: Return the answer identified by the given `answer_id`.
//
//   * The user should have at least 'content' access rights to the `attempts.item_id` item for
//     `answers.attempt_id`.
//
//   * The user should be able to see answers related to his group's attempts so
//     the user should be a member of the `attempts.group_id` team or
//     `attempts.group_id` should be equal to the user's self group.
// parameters:
// - name: answer_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/itemAnswerGetResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) get(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	answerID, err := service.ResolveURLQueryPathInt64Field(httpReq, "answer_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	var result []map[string]interface{}
	err = srv.Store.Answers().Visible(user).
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Where("answers.id = ?", answerID).
		Select(`answers.id, answers.author_id, attempts.item_id, answers.attempt_id,
			answers.type, answers.state, answers.answer, answers.created_at, gradings.score,
			gradings.graded_at`).
		ScanIntoSliceOfMaps(&result).Error()
	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}
