package items

import (
	"net/http"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /items/{item_id}/participant/{participant_id}/thread/log items itemActivityLogForThread
//
//	---
//	summary: Activity log on an item and participant for a thread
//	description: >
//		Returns rows from `answers` and started/validated `results`
//		with additional info on users and items for the `{participant_id}` group.
//
//
//		If possible, item titles are shown in the authenticated user's default language.
//		Otherwise, the item's default language is used.
//
//
//		`first_name` and `last_name` (from `profile`) of users are only visible to the users themselves and
//		to managers of those users' groups to which they provided view access to personal data.
//
//
//		All rows of the result are related to the `{participant_id}` group
//		and the `{item_id}` item.
//
//
//			- The thread for the `{participant_id}`-`{item_id}` pair should exist AND
//
//			- The current user or one of his teams should be allowed to view the `{item_id}` item (`can_view` >= 'content) AND
//
//			- One of the following:
//				- the current user should be a member of the `{participant_id}` team or
//				  `{participant_id}` should be equal to the user's self group OR
//				- the current user should have `can_watch` >= 'answer' permission on the `{item_id}` OR
//				- the current user should have `can_watch` >= 'result' permission on the thread's item, AND
//					the current user should be a descendant of the group the participant has requested help from in the given thread, AND
//		      the thread should be either open or closed for less than 2 weeks, AND
//		      the current user (personally or as a team) should have a validated result on `{item_id}`.
//
//			If any of the preconditions fails, the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: participant_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: from.attempt_id
//			description: Start the page from the row next to the row with `attempt_id`=`{from.attempt_id}`
//							 (all other `{from.*}` parameters are required when `{from.attempt_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.answer_id
//			description: Start the page from the row next to the row with `from_answer_id`=`{from.answer_id}`
//							 (all other `{from.*}` parameters are required when `{from.answer_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.activity_type
//			description: Start the page from the row next to the row with `activity_type`=`{from.activity_type}`
//							 (all other `{from.*}` parameters are required when `{from.activity_type}` is present)
//			in: query
//			type: string
//			enum: [result_started,submission,result_validated,saved_answer,current_answer]
//		- name: limit
//			description: Display the first N rows
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array with the activity log
//			schema:
//				type: array
//				items:
//			 		"$ref": "#/definitions/itemActivityLogResponseRow"
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
func (srv *Service) getActivityLogForThread(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)

	store := srv.GetStore(httpRequest)

	userAndHisTeamsQuery := store.Raw("SELECT id FROM ? `teams` UNION ALL SELECT ?",
		store.ActiveGroupGroups().
			WhereUserIsMember(user).
			Where("groups_groups_active.is_team_membership = 1").
			Select("groups_groups_active.parent_group_id AS id").SubQuery(),
		user.GroupID)

	// the current user has at least 'content' access on the threads.item_id
	userHasViewContentPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("view", "content").
		Where("permissions.item_id = threads.item_id").
		Select("1").Limit(1).SubQuery()
	// a team of the current user has at least 'content' access on the threads.item_id
	userTeamHasViewContentPermOnItemSubQuery := store.Permissions().
		Joins("JOIN `groups_ancestors_active` ON groups_ancestors_active.ancestor_group_id = permissions.group_id").
		Joins("JOIN `groups_groups_active` ON groups_groups_active.parent_group_id = groups_ancestors_active.child_group_id").
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Where("groups_groups_active.is_team_membership = 1").
		WherePermissionIsAtLeast("view", "content").
		Where("permissions.item_id = threads.item_id").
		Select("1").Limit(1).SubQuery()

	// the current user has 'can_watch'>='answer' permission on the threads.item_id
	userHasCanWatchAnswerPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "answer").
		Where("permissions.item_id = threads.item_id").
		Select("1").Limit(1).SubQuery()
	// the current user has 'can_watch'>='result' permission on the threads.item_id
	userHasCanWatchResultPermOnItemSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "result").
		Where("permissions.item_id = threads.item_id").
		Select("1").Limit(1).SubQuery()
	// the current user or his team has a validated result on the threads.item_id
	userOrHisTeamHasValidatedResultOnItemSubQuery := store.Results().
		Where("results.item_id = threads.item_id").
		Where("results.validated").
		Where("results.participant_id IN (SELECT id FROM user_and_his_teams)").
		Select("1").Limit(1).SubQuery()
	// the current user is a descendant of the thread's helper group and the thread is active or closed less than 2 weeks ago
	userIsAHelperAndTheThreadHasNotBeenExpiredSubQuery := store.ActiveGroupAncestors().
		Where("child_group_id = ?", user.GroupID).
		Where("ancestor_group_id = threads.helper_group_id").
		Where("threads.status IN ('waiting_for_participant', 'waiting_for_trainer') OR threads.latest_update_at > NOW() - INTERVAL 2 WEEK").
		Select("1").Limit(1).SubQuery()

	var found []struct{}
	err = store.Threads().
		Select("1").
		Where("threads.participant_id = ?", participantID).
		Where("threads.item_id = ?", itemID).
		With("user_and_his_teams", userAndHisTeamsQuery).
		Where(`
				(? OR ?) AND
				(
					threads.participant_id IN (SELECT id from user_and_his_teams) OR
					? OR
					(? AND ? AND ?)
				)`,
			/* ( */
			userHasViewContentPermOnItemSubQuery /* OR */, userTeamHasViewContentPermOnItemSubQuery,
			/* ) */
			/* AND */
			/* ( */
			/*   [the user/(his team) is the participant] */
			/*   OR */
			userHasCanWatchAnswerPermOnItemSubQuery,
			/*   OR */
			/*   ( */
			userHasCanWatchResultPermOnItemSubQuery,
			/*	   AND */
			userIsAHelperAndTheThreadHasNotBeenExpiredSubQuery,
			/*	   AND */
			userOrHisTeamHasValidatedResultOnItemSubQuery,
			/*   ) */
			/* ) */).
		Limit(1).
		Scan(&found).Error()

	service.MustNotBeError(err)
	if len(found) == 0 {
		return service.ErrAPIInsufficientAccessRights
	}

	return srv.getActivityLog(
		responseWriter, httpRequest,
		[]string{"activity_type", "attempt_id", "answer_id"},
		func(query *database.DB) *database.DB { return query },
		func(answersQuery *database.DB) *database.DB {
			return answersQuery.
				Where("answers.item_id = ?", itemID).
				Where("answers.participant_id = ?", participantID)
		},
		nil,
		func(startedResultsQuery *database.DB) *database.DB {
			return startedResultsQuery.
				Where("started_results.item_id = ?", itemID).
				Where("started_results.participant_id = ?", participantID)
		},
		func(validatedResultsQuery *database.DB) *database.DB {
			return validatedResultsQuery.
				Where("validated_results.item_id = ?", itemID).
				Where("validated_results.participant_id = ?", participantID)
		},
		func(unQuery *database.DB) *database.DB {
			if participantID == user.GroupID {
				return unQuery.With("can_watch_answer", store.Raw("SELECT 1 AS can_watch_answer"))
			}
			return unQuery.With("can_watch_answer",
				store.Raw("SELECT EXISTS(?) OR (EXISTS(?) AND EXISTS(?)) AS can_watch_answer",
					store.ActiveGroupGroups().WhereUserIsMember(user).
						Where("groups_groups_active.is_team_membership").
						Where("groups_groups_active.parent_group_id = ?", participantID).
						Select("1").Limit(1).QueryExpr(),
					store.ActiveGroupAncestors().ManagedByUser(user).
						Where("group_managers.can_watch_members").
						Where("groups_ancestors_active.child_group_id = ?", participantID).
						Select("1").Limit(1).QueryExpr(),
					store.Permissions().
						Joins("JOIN groups_ancestors_active AS ancestors ON ancestors.ancestor_group_id = permissions.group_id").
						Where("ancestors.child_group_id = ?", user.GroupID).
						WherePermissionIsAtLeast("watch", "answer").
						Where("permissions.item_id = ?", itemID).
						Select("1").Limit(1).QueryExpr(),
				))
		},
		"(SELECT can_watch_answer FROM can_watch_answer) AS can_watch_answer",
		"-at,-attempt_id",
		false)
}
