package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /items/{item_id}/best-answer answers bestAnswerGet
// ---
// summary: Get the best answer
// description: Returns the best answer of the user on the given `{item_id}` among all attempts. The best answer is
//              defined as the one which gives the highest score. If there are several, it is the most recent one
//							by `created_at`.
//
//   * if `watched_group_id` is given, the current user must be allowed to watch "answer" of the item identified by `item_id`
//   * if `watched_group_id` is given, it must be a participant (= team or user)
//   * if `watched_group_id` is given, the current user must be allowed to watch the participant
//
//
//   If any of the preconditions fails, the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: watched_group_id
//   in: query
//   type: integer
//   description: A participant (`team_id` or user). If given, get the best answer of the participant instead
//				  		  of the one of the user.
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
func (srv *Service) getBestAnswer(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	watchedGroupID, watchedGroupIDSet, apiError := srv.ResolveWatchedGroupID(httpReq)
	if apiError != service.NoError {
		return apiError
	}

	user := srv.GetUser(httpReq)
	store := srv.GetStore(httpReq)
	var result []map[string]interface{}

	var bestAnswerQuery *database.DB
	if watchedGroupID != 0 && watchedGroupIDSet {
		// check 'can_watch'>='answer' permission on the answers.item_id
		itemPerms := store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("watch", "answer").
			Where("permissions.item_id = answers.item_id").
			Select("1").
			Limit(1)

		// check if able to watch the participant
		participantPerms := store.ActiveGroupAncestors().ManagedByUser(user).
			Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id").
			Where("groups_ancestors_active.child_group_id = answers.participant_id").
			Where("can_watch_members").
			Select("1").
			Limit(1)

		bestAnswerQuery = store.Answers().
			Where(`?`, itemPerms.SubQuery()).
			Where(`?`, participantPerms.SubQuery()).
			Where("participant_id = ?", watchedGroupID)
	} else {
		bestAnswerQuery = store.Answers().
			Visible(user).
			Where("author_id = ?", user.GroupID)
	}

	err = withGradings(bestAnswerQuery).
		Where("item_id = ?", itemID).
		Where("`type` = 'Submission'").
		Where("graded_at IS NOT NULL").
		Order("gradings.score DESC").
		Order("answers.created_at DESC").
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
