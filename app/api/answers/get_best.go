package answers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /items/{item_id}/best-answer answers bestAnswerGet
//
//	---
//	summary: Get the best answer
//	description: >
//		Returns the best answer of the user on the given `item_id` among all attempts.
//		The best answer is defined as the one which gives the highest score.
//		If there are several, it is the most recent one by `created_at`.
//
//		* if `watched_group_id` is given, the current user must be allowed to watch "answer" of the item identified by `item_id`
//		* if `watched_group_id` is given, it must be a participant (= team or user)
//		* if `watched_group_id` is given, the current user must be allowed to watch the participant
//
//		If any of the preconditions fails, the 'forbidden' error is returned.
//
//		If the preconditions pass but there is no answer for the user on the given `item_id`, returns a `404`.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: watched_group_id
//			in: query
//			type: integer
//			format: int64
//			description: >
//				A participant (`team_id` or user).
//				If given, get the best answer of the participant instead of the one of the user.
//	responses:
//		"200":
//			"$ref": "#/responses/itemAnswerGetResponse"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"404":
//			"$ref": "#/responses/notFoundResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getBestAnswer(rw http.ResponseWriter, httpReq *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	watchedGroupID, watchedGroupIDIsSet, err := srv.ResolveWatchedGroupID(httpReq)
	service.MustNotBeError(err)

	user := srv.GetUser(httpReq)
	store := srv.GetStore(httpReq)
	var result []map[string]interface{}

	bestAnswerQuery := store.Answers().
		WithGradings().
		DB

	if watchedGroupIDIsSet {
		// the following checks were made by ResolveWatchedGroupID:
		// - watched_group_id must be a participant
		// - the current user is able to watch the participant

		// check 'can_watch'>='answer' permission on the answers.item_id
		if !user.CanWatchItemAnswer(store, itemID) {
			return service.InsufficientAccessRightsError
		}

		bestAnswerQuery = bestAnswerQuery.
			Where("participant_id = ?", watchedGroupID)
	} else {
		// check 'can_view'>='content' permission on the answers.item_id
		if !user.CanViewItemContent(store, itemID) {
			return service.InsufficientAccessRightsError
		}

		bestAnswerQuery = bestAnswerQuery.
			Where("participant_id = ?", user.GroupID)
	}

	err = bestAnswerQuery.
		Where("item_id = ?", itemID).
		Where("`type` = 'Submission'").
		Where("graded_at IS NOT NULL").
		Order("gradings.score DESC").
		Order("answers.created_at DESC").
		Limit(1).
		ScanIntoSliceOfMaps(&result).Error()
	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.ErrNotFound(errors.New("no answer found"))
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return nil
}
