package contests

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /contests/{item_id}/enter contests contestEnter
// ---
// summary: Enter the contest
// description: >
//                Allows to enter a contest as a user or as a team (if `as_team_id` is given).
//
//
//                Restrictions:
//                  * `item_id` should be a contest;
//                  * `as_team_id` (if given) should be the current user's team;
//                  * the authenticated user (or his team) should have at least 'info' access to the item;
//                  * the group (the user or his team) must be qualified for the contest (contestGetQualificationState returns "ready").
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a contest"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
//   format: int64
// responses:
//   "201":
//     "$ref": "#/responses/contestEnterResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) enter(w http.ResponseWriter, r *http.Request) service.APIError {
	apiError := service.NoError
	var qualificationState *contestGetQualificationStateResponse
	var itemInfo struct {
		Now                        *database.Time
		Duration                   *string
		ContestParticipantsGroupID *int64
	}
	err := srv.Store.InTransaction(func(store *database.DataStore) error {
		qualificationState, apiError = srv.getContestInfoAndQualificationStateFromRequest(r, store, true)
		if apiError != service.NoError {
			return apiError.Error
		}

		if qualificationState.State != string(ready) {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error
		}

		service.MustNotBeError(store.Items().ByID(qualificationState.itemID).
			Select("NOW() AS now, items.duration, items.contest_participants_group_id").
			WithWriteLock().Take(&itemInfo).Error())

		var attemptID int64
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			service.MustNotBeError(retryStore.Attempts().
				Where("participant_id = ?", qualificationState.groupID).
				WithWriteLock().
				PluckFirst("IFNULL(MAX(id), 0) + 1", &attemptID).Error())

			user := srv.GetUser(r)
			return retryStore.Attempts().InsertMap(map[string]interface{}{
				"id": attemptID, "participant_id": qualificationState.groupID, "created_at": itemInfo.Now,
				"creator_id": user.GroupID, "parent_attempt_id": 0, "root_item_id": qualificationState.itemID,
			})
		}))
		service.MustNotBeError(store.Results().InsertMap(map[string]interface{}{
			"attempt_id": attemptID, "participant_id": qualificationState.groupID,
			"item_id": qualificationState.itemID, "started_at": itemInfo.Now,
		}))

		if itemInfo.ContestParticipantsGroupID != nil {
			var totalAdditionalTime int64
			service.MustNotBeError(store.ActiveGroupAncestors().
				Where("groups_ancestors_active.child_group_id = ?", qualificationState.groupID).
				Joins(`
					LEFT JOIN groups_contest_items
						ON groups_contest_items.group_id = groups_ancestors_active.ancestor_group_id AND
							groups_contest_items.item_id = ?`, qualificationState.itemID).
				Group("groups_ancestors_active.child_group_id").
				WithWriteLock().
				PluckFirst("IFNULL(SUM(TIME_TO_SEC(groups_contest_items.additional_time)), 0)", &totalAdditionalTime).
				Error())
			service.MustNotBeError(store.Exec(`
				INSERT INTO groups_groups (parent_group_id, child_group_id, expires_at)
				VALUES(?, ?, IFNULL(DATE_ADD(?, INTERVAL (TIME_TO_SEC(?) + ?) SECOND), '9999-12-31 23:59:59'))
				ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
				itemInfo.ContestParticipantsGroupID, qualificationState.groupID,
				itemInfo.Now, itemInfo.Duration, totalAdditionalTime).Error())
			service.MustNotBeError(store.GroupGroups().After())
			// Upserting into groups_groups may mark some attempts as 'to_be_propagated',
			// so we need to recompute them
			service.MustNotBeError(store.Attempts().ComputeAllAttempts())
		} else {
			logging.GetLogEntry(r).Warnf("items.contest_participants_group_id is not set for the item with id = %d", qualificationState.itemID)
		}

		return nil
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	service.MustNotBeError(render.Render(w, r, service.CreationSuccess(map[string]interface{}{
		"duration":   itemInfo.Duration,
		"entered_at": itemInfo.Now,
	})))
	return service.NoError
}
