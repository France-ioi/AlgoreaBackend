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
//   * The user should have at least partial access rights to the `users_answers.idItem` item
//
//   * The user should be able to see answers related to his group's attempts, so
//      (a) if `items.bHasAttempts = 1`, then the user should be a member of the groups_attempts.idGroup team,
//      (b) if `items.bHasAttempts = 0`, then groups_attempts.idGroup should be equal to the user's self group
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
	userAnswerID, err := service.ResolveURLQueryPathInt64Field(httpReq, "answer_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	var result []map[string]interface{}
	err = srv.Store.UserAnswers().Visible(user).
		Where("users_answers.ID = ?", userAnswerID).
		Select(`users_answers.ID, users_answers.idUser, users_answers.idItem, users_answers.idAttempt,
			users_answers.sType, users_answers.sState, users_answers.sAnswer,
			users_answers.sSubmissionDate, users_answers.iScore, users_answers.bValidated,
			users_answers.sGradingDate, users_answers.idUserGrader`).
		ScanIntoSliceOfMaps(&result).Error()
	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}
