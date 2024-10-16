package answers

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /answers/{answer_id} answers answerGet
//
//		---
//		summary: Get an answer
//		description: >
//			Returns the answer identified by the given `{answer_id}`.
//
//			- If the user is a participant
//				- (s)he (or one of his/her teams) should have at least 'content' access rights to the `answers.item_id` and
//				- be a member of the `answers.participant_id` team or
//				  `answers.participant_id` should be equal to the user's self group.
//
//			- If the user is an observer (a manager with `can_watch_members` of an ancestor of `answers.participant_id` group)
//				- (s)he should have `can_watch` >= 'answer' permission on the `answers.item_id` OR
//				- `can_watch` >= 'result' permission on the `answers.item_id` together with a validated result
//	        (personally or as a team) on the `answers.item_id`.
//
//			- If the user is a thread reader (when the thread for the `answers.participant_id`-`answers.item_id` pair exists)
//				- (s)he should have `can_watch` >= 'answer' permission on the `answers.item_id` OR
//				- (s)he should be a descendant of the thread's `helper_group_id` and have `can_watch` >= 'result' permission
//				  on the `answers.item_id` together with a validated result (personally or as a team) on the `answers.item_id`
//				  while the thread should be active or closed less than 2 weeks ago.
//
//			If any of the preconditions fails, the 'forbidden' error is returned.
//		parameters:
//			- name: answer_id
//				in: path
//				type: integer
//				required: true
//				format: int64
//		responses:
//			"200":
//				"$ref": "#/responses/itemAnswerGetResponse"
//			"400":
//				"$ref": "#/responses/badRequestResponse"
//			"401":
//				"$ref": "#/responses/unauthorizedResponse"
//			"403":
//				"$ref": "#/responses/forbiddenResponse"
//			"500":
//				"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAnswer(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	answerID, err := service.ResolveURLQueryPathInt64Field(httpReq, "answer_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	var result []map[string]interface{}

	store := srv.GetStore(httpReq)

	userAndHisTeamsQuery := store.Raw("SELECT id FROM ? `teams` UNION ALL SELECT ?",
		store.ActiveGroupGroups().
			WhereUserIsMember(user).
			Joins("JOIN `groups` ON groups.id = groups_groups_active.parent_group_id AND groups.type='Team'").
			Select("groups.id").SubQuery(),
		user.GroupID)

	// a participant should have at least 'content' access to the answers.item_id
	userHasViewContentPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("view", "content").
		Where("permissions.item_id = answers.item_id").
		Select("1").Limit(1).SubQuery()
	// or a participant's team should have at least 'content' access to the answers.item_id
	userTeamHasViewContentPermOnItemSubQuery := store.Permissions().
		Joins("JOIN `groups_ancestors_active` ON groups_ancestors_active.ancestor_group_id = permissions.group_id").
		Joins("JOIN `groups_groups_active` ON groups_groups_active.parent_group_id = groups_ancestors_active.child_group_id").
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Joins("JOIN `groups` ON groups.id = groups_groups_active.parent_group_id AND groups.type='Team'").
		WherePermissionIsAtLeast("view", "content").
		Where("permissions.item_id = answers.item_id").
		Select("1").Limit(1).SubQuery()

	// an observer/thread viewer should have 'can_watch'>='answer' permission on the answers.item_id
	userHasCanWatchAnswerPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "answer").
		Where("permissions.item_id = answers.item_id").
		Select("1").Limit(1).SubQuery()
	// or an observer/helper should have 'can_watch'>='result' permission on the answers.item_id
	userHasCanWatchResultPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "result").
		Where("permissions.item_id = answers.item_id").
		Select("1").Limit(1).SubQuery()
	// and an observer/helper or his team should have a validated result on the answers.item_id
	userOrHisTeamHasValidatedResultOnItemSubQuery := store.Results().
		Where("results.item_id = answers.item_id").
		Where("results.validated").
		Where("results.participant_id IN (SELECT id FROM user_and_his_teams)").
		Select("1").Limit(1).SubQuery()

	// an observer should be able to watch the participant
	userIsAManagerThatCanWatchMembersSubQuery := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = answers.participant_id").
		Where("can_watch_members").
		Select("1").Limit(1).SubQuery()

	// the thread should exist to allow thread viewers with 'can_watch'>='answer' permission to view the answer
	theThreadExistsSubQuery := store.Threads().
		Where("threads.participant_id = answers.participant_id").
		Where("threads.item_id = answers.item_id").
		Select("1").Limit(1).SubQuery()

	// a helper should be an ancestor of the thread's helper group and the thread should be active or closed less than 2 weeks ago
	userIsAHelperAndTheThreadHasNotBeenExpiredSubQuery := store.Threads().
		Where("threads.participant_id = answers.participant_id").
		Where("threads.item_id = answers.item_id").
		Where("threads.status IN ('waiting_for_participant', 'waiting_for_trainer') OR threads.latest_update_at > NOW() - INTERVAL 2 WEEK").
		Joins(`
			JOIN groups_ancestors_active
				ON groups_ancestors_active.child_group_id = ? AND groups_ancestors_active.ancestor_group_id = threads.helper_group_id`,
			user.GroupID).
		Select("1").Limit(1).SubQuery()

	err = store.Answers().
		WithGradings().
		ByID(answerID).
		With("user_and_his_teams", userAndHisTeamsQuery).
		// 1) the user is the participant or a member of the participant team able to view the item,
		// 2) or an observer with required permissions
		// 3) or a thread viewer with required permissions
		Where(`
				((? OR ?) AND answers.participant_id IN (SELECT id from user_and_his_teams)) OR
				(? AND (? OR ?)) OR
				(? AND ? AND (? OR ?))`,
			/* ( */
			/*   ( */
			userHasViewContentPermOnItemSubQuery /* OR */, userTeamHasViewContentPermOnItemSubQuery,
			/*   ) */
			/*   AND [the user/(his team) is the participant] */
			/* ) */
			/* OR */
			/* ( */
			userHasCanWatchAnswerPermOnItemSubQuery,
			/*   AND */
			/*   ( */
			userIsAManagerThatCanWatchMembersSubQuery /* OR */, theThreadExistsSubQuery,
			/*   ) */
			/* ) */
			/* OR */
			/* ( */
			userHasCanWatchResultPermOnItemSubQuery,
			/*   AND */
			userOrHisTeamHasValidatedResultOnItemSubQuery,
			/*   AND */
			/*   ( */
			userIsAManagerThatCanWatchMembersSubQuery /* OR */, userIsAHelperAndTheThreadHasNotBeenExpiredSubQuery,
			/* ) */).
		ScanIntoSliceOfMaps(&result).Error()

	service.MustNotBeError(err)
	if len(result) == 0 {
		return service.InsufficientAccessRightsError
	}
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)[0]

	render.Respond(rw, httpReq, convertedResult)
	return service.NoError
}
