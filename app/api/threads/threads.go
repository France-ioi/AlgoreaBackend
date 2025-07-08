// Package threads provides API services for threads managing.
package threads

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `items`.
type Service struct {
	*service.Base
}

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))

	router.Get("/threads", service.AppHandler(srv.listThreads).ServeHTTP)
	router.Get("/items/{item_id}/participant/{participant_id}/thread", service.AppHandler(srv.getThread).ServeHTTP)
	router.Put("/items/{item_id}/participant/{participant_id}/thread", service.AppHandler(srv.updateThread).ServeHTTP)
}

type threadInfo struct {
	ThreadStatus                  string
	ThreadHelperGroupID           int64
	ThreadMessageCount            int
	ThreadIsOpen                  bool
	ThreadWasUpdatedRecently      bool
	UserCanWatchForParticipant    bool
	UserCanWatchAnswer            bool
	UserCanWatchResult            bool
	UserIsDescendantOfHelperGroup bool
	UserHasValidatedResultOnItem  bool
}

func userCanWriteInThread(user *database.User, participantID int64, threadInfo *threadInfo) bool {
	return database.IsThreadOpenStatus(threadInfo.ThreadStatus) &&
		((participantID == user.GroupID) ||
			(threadInfo.UserCanWatchAnswer && threadInfo.UserCanWatchForParticipant) ||
			(threadInfo.UserIsDescendantOfHelperGroup &&
				(threadInfo.UserCanWatchAnswer ||
					threadInfo.UserCanWatchResult && threadInfo.UserHasValidatedResultOnItem)))
}

func userCanWatchForThread(threadInfo *threadInfo) bool {
	return threadInfo.UserCanWatchAnswer && threadInfo.UserCanWatchForParticipant
}

func constructThreadInfoQuery(store *database.DataStore, user *database.User, itemID, participantID int64) *database.DB {
	canWatchForParticipantSubQuery := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = ?", participantID).
		Where("group_managers.can_watch_members").
		Select("1").Limit(1).SubQuery()

	canWatchAnswerSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "answer").
		Where("permissions.item_id = ?", itemID).
		Select("1").Limit(1).SubQuery()

	canWatchResultSubQuery := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "result").
		Where("permissions.item_id = ?", itemID).
		Select("1").Limit(1).SubQuery()

	userIsDescendantOfHelperGroupSubQuery := store.ActiveGroupAncestors().
		Where("groups_ancestors_active.ancestor_group_id = threads.helper_group_id").
		Where("groups_ancestors_active.child_group_id = ?", user.GroupID).
		Select("1").Limit(1).SubQuery()

	userHasValidatedResultOnItemSubQuery := store.Results().
		Where("results.item_id = threads.item_id").
		Where("results.validated").
		Where("results.participant_id = ?", user.GroupID).
		Select("1").Limit(1).SubQuery()

	return store.Table("(SELECT 1) AS t").
		Joins("LEFT JOIN threads ON threads.participant_id = ? AND threads.item_id = ?", participantID, itemID).
		Select(`
			IFNULL(threads.status, 'not_started') AS thread_status,
			threads.latest_update_at AS thread_latest_update_at,
			threads.helper_group_id AS thread_helper_group_id,
			threads.message_count AS thread_message_count,
			threads.status IN ('waiting_for_participant', 'waiting_for_trainer') AS thread_is_open,
			threads.latest_update_at > NOW() - INTERVAL 2 WEEK AS thread_was_updated_recently,
			? AS user_can_watch_for_participant,
			? AS user_can_watch_answer,
			? AS user_can_watch_result,
			? AS user_is_descendant_of_helper_group,
			? AS user_has_validated_result_on_item`,
			canWatchForParticipantSubQuery,
			canWatchAnswerSubQuery,
			canWatchResultSubQuery,
			userIsDescendantOfHelperGroupSubQuery,
			userHasValidatedResultOnItemSubQuery)
}
