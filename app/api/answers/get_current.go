package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /items/{item_id}/current-answer answers currentAnswerGet
// ---
// summary: Get a current answer
// description: Returns the latest auto-saved ('Current') answer for the given `{item_id}` and `{attempt_id}`.
//
//   * The user should have at least 'content' access rights to the `item_id` item.
//
//   * The user should be able to see answers related to his group's attempts so
//     the user should be a member of the `answers.participant_id` team or
//     `answers.participant_id` should be equal to the user's self group.
//
//   * `{as_team_id}` (if given) should be the user's team.
//
//
//   If any of the preconditions fails, the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: attempt_id
//   in: query
//   type: integer
//   format: int64
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
//   format: int64
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

	user := srv.GetUser(httpReq)
	var result []map[string]interface{}
	err = visibleAnswersWithGradings(srv.Store, user).
		Where("participant_id = ?", participantID).
		Where("attempt_id = ?", attemptID).
		Where("item_id = ?", itemID).
		Where("type = 'Current'").
		Order("created_at DESC").
		Limit(1).
		ScanIntoSliceOfMaps(&result).Error()
	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}

func visibleAnswersWithGradings(store *database.DataStore, user *database.User) *database.DB {
	return withGradings(store.Answers().Visible(user))
}

func withGradings(answersQuery *database.DB) *database.DB {
	return answersQuery.
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Select(`answers.id, answers.author_id, answers.item_id, answers.attempt_id, answers.participant_id,
			answers.type, answers.state, answers.answer, answers.created_at, gradings.score,
			gradings.graded_at`)
}
