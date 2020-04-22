package contests

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /attempts/{attempt_id}/items/{item_id}/enter items itemEnter
// ---
// summary: Enter the item
// description: >
//                Allows to enter an item requiring explicit entry as a user or as a team (if `as_team_id` is given).
//
//
//                Restrictions:
//                  * `item_id` should require explicit entry;
//                  * `as_team_id` (if given) should be the current user's team;
//                  * an attempt with `participant_id` = `as_team_id` (or the current user) and `id` = `attempt_id` should exist;
//                  * the authenticated user (or his team) should have at least 'info' access to the item;
//                  * the group (the user or his team) must be qualified for the item (itemGetEntryState returns "ready").
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: attempt_id
//   description: "`id` of an attempt which will be used as a parent attempt for the participation"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: item_id
//   description: "`id` of an item"
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
//     "$ref": "#/responses/itemEnterResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) enter(w http.ResponseWriter, r *http.Request) service.APIError {
	parentAttemptID, err := service.ResolveURLQueryPathInt64Field(r, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	apiError := service.NoError
	var entryState *itemGetEntryStateResponse
	var itemInfo struct {
		Now                 *database.Time
		Duration            *string
		ParticipantsGroupID *int64
	}
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		entryState, apiError = srv.getItemInfoAndEntryStateFromRequest(r, store, true)
		if apiError != service.NoError {
			return apiError.Error
		}

		if entryState.State != string(ready) {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error
		}

		service.MustNotBeError(store.Items().ByID(entryState.itemID).
			Select("NOW() AS now, items.duration, items.participants_group_id").
			WithWriteLock().Take(&itemInfo).Error())

		var totalAdditionalTime int64
		service.MustNotBeError(store.ActiveGroupAncestors().
			Where("groups_ancestors_active.child_group_id = ?", entryState.groupID).
			Joins(`
					LEFT JOIN groups_contest_items
						ON groups_contest_items.group_id = groups_ancestors_active.ancestor_group_id AND
							groups_contest_items.item_id = ?`, entryState.itemID).
			Group("groups_ancestors_active.child_group_id").
			WithWriteLock().
			PluckFirst("IFNULL(SUM(TIME_TO_SEC(groups_contest_items.additional_time)), 0)", &totalAdditionalTime).
			Error())

		user := srv.GetUser(r)
		service.MustNotBeError(store.Attempts().InsertMap(map[string]interface{}{
			"id": gorm.Expr("(SELECT * FROM ? AS max_attempt)", store.Attempts().Select("IFNULL(MAX(id)+1, 0)").
				Where("participant_id = ?", entryState.groupID).WithWriteLock().SubQuery()),
			"participant_id": entryState.groupID, "created_at": itemInfo.Now,
			"creator_id": user.GroupID, "parent_attempt_id": parentAttemptID, "root_item_id": entryState.itemID,
			"allows_submissions_until": gorm.Expr("IFNULL(DATE_ADD(?, INTERVAL (TIME_TO_SEC(?) + ?) SECOND), '9999-12-31 23:59:59')",
				(*time.Time)(itemInfo.Now), itemInfo.Duration, totalAdditionalTime),
		}))
		var attemptID int64
		service.MustNotBeError(store.Attempts().
			Where("participant_id = ?", entryState.groupID).
			Where("parent_attempt_id = ?", parentAttemptID).
			PluckFirst("MAX(id)", &attemptID).Error())

		service.MustNotBeError(store.Results().InsertMap(map[string]interface{}{
			"attempt_id": attemptID, "participant_id": entryState.groupID,
			"item_id": entryState.itemID, "started_at": itemInfo.Now,
		}))

		if itemInfo.ParticipantsGroupID != nil {
			service.MustNotBeError(store.Exec(`
				INSERT INTO groups_groups (parent_group_id, child_group_id, expires_at)
				VALUES(?, ?, IFNULL(DATE_ADD(?, INTERVAL (TIME_TO_SEC(?) + ?) SECOND), '9999-12-31 23:59:59'))
				ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
				itemInfo.ParticipantsGroupID, entryState.groupID,
				itemInfo.Now, itemInfo.Duration, totalAdditionalTime).Error())
			service.MustNotBeError(store.GroupGroups().After())
			// Upserting into groups_groups may mark some attempts as 'to_be_propagated',
			// so we need to recompute them
			service.MustNotBeError(store.Results().Propagate())
		} else {
			logging.GetLogEntry(r).Warnf("items.participants_group_id is not set for the item with id = %d", entryState.itemID)
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
