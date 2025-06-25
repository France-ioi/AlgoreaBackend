package items

import (
	"net/http"
	"sync/atomic"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// The request has successfully updated the object
// swagger:response updatedStartResultResponse
type updatedStartResultResponse struct { //nolint:unused
	// in: body
	Body struct {
		// "updated"
		// enum: updated
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data attemptsListResponseRow `json:"data"`
	}
}

// swagger:operation POST /items/{ids}/start-result items resultStart
//
//		---
//		summary: Start a result
//		description: >
//			Creates a new started result for the given item and attempt or sets `started_at` of an existing result (if it hasn't been set).
//	   The started result is then returned as `data`.
//			If `as_team_id` is given, the created result is linked to the `as_team_id` group instead of the user's self group.
//
//
//				Restrictions:
//
//			* if `as_team_id` is given, it should be a user's parent team group,
//			* the first item in `{ids}` should be a root activity/skill (groups.root_activity_id/root_skill_id) of a group
//				the participant is a descendant of or manages,
//			* the final item in `{ids}` should not require explicit entry (`items.requires_explicit_entry` should be false),
//			* `{ids}` should be an ordered list of parent-child items,
//			* the group starting the result should have at least 'content' access on each of the items in `{ids}`,
//			* the participant should have a started, allowing submission, not ended result for each item but the last,
//				with `{attempt_id}` (or its parent attempt each time we reach a root of an attempt) as the attempt,
//			* if `{ids}` consists of only one item, the `{attempt_id}` should be zero,
//
//			otherwise the 'forbidden' error is returned.
//		parameters:
//			- name: ids
//				in: path
//				type: string
//				description: slash-separated list of item IDs (no more than 10 IDs)
//				required: true
//			- name: attempt_id
//				in: query
//				type: integer
//				format: int64
//				required: true
//			- name: as_team_id
//				in: query
//				type: integer
//				format: int64
//		responses:
//			"200":
//				"$ref": "#/responses/updatedStartResultResponse"
//			"400":
//				"$ref": "#/responses/badRequestResponse"
//			"401":
//				"$ref": "#/responses/unauthorizedResponse"
//			"403":
//				"$ref": "#/responses/forbiddenResponse"
//			"408":
//				"$ref": "#/responses/requestTimeoutResponse"
//			"500":
//				"$ref": "#/responses/internalErrorResponse"
func (srv *Service) startResult(w http.ResponseWriter, r *http.Request) error {
	var err error

	ids, err := idsFromRequest(r)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	attemptID, err := service.ResolveURLQueryGetInt64Field(r, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID := service.ParticipantIDFromContext(r.Context())

	var attemptInfo attemptsListResponseRow
	var shouldSchedulePropagation bool
	store := srv.GetStore(r)
	err = store.InTransaction(func(store *database.DataStore) error {
		var ok bool
		ok, err = store.Items().IsValidParticipationHierarchyForParentAttempt(ids, participantID, attemptID, true, true)
		service.MustNotBeError(err)
		if !ok {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		onBeforeInsertingResultInResultStartHook.Load().(func())()

		itemID := ids[len(ids)-1]
		var found bool
		found, err = store.Items().ByID(itemID).
			Where("NOT items.requires_explicit_entry").WithSharedWriteLock().HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		result := store.Exec(`
			INSERT INTO results (participant_id, attempt_id, item_id, started_at, latest_activity_at)
			VALUES (?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				latest_activity_at = IF(started_at IS NULL, NOW(), latest_activity_at),
				started_at = IFNULL(started_at, NOW())`,
			participantID, attemptID, itemID)
		service.MustNotBeError(result.Error())

		if result.RowsAffected() != 0 {
			resultStore := store.Results()
			service.MustNotBeError(resultStore.MarkAsToBePropagated(participantID, attemptID, itemID, false))
			shouldSchedulePropagation = true
		}

		service.MustNotBeError(constructQueryForGettingAttemptsList(store, participantID, itemID, srv.GetUser(r)).
			Where("attempts.id = ?", attemptID).
			WithCustomWriteLocks(golang.NewSet("results"), golang.NewSet[string]()).
			Scan(&attemptInfo).Error())

		if attemptInfo.UserCreator.GroupID == nil {
			attemptInfo.UserCreator = nil
		}

		return nil
	})
	service.MustNotBeError(err)

	if shouldSchedulePropagation {
		service.SchedulePropagation(store, srv.GetPropagationEndpoint(), []string{"results"})
	}

	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(&attemptInfo)))
	return nil
}

var onBeforeInsertingResultInResultStartHook atomic.Value

func init() { //nolint:gochecknoinits // this is an initialization function to store the default hook
	onBeforeInsertingResultInResultStartHook.Store(func() {})
}
