package contests

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation POST /contests/{item_id}/groups/{group_id} contests contestEnter
// ---
// summary: Enter the contest
// description: >
//                Allows to enter a contest as a group (user self or team).
//
//
//                Restrictions:
//                  * `item_id` should be a contest;
//                  * `group_id` should be either the current user's self group (if the item's `allows_multiple_attempts` is false) or
//                     a team with `team_item_id` = `item_id` (otherwise);
//                  * the authenticated user should have at least 'info' access to the item;
//                  * the authenticated user should be a member of the `group_id` (if it is a team);
//                  * the group must be qualified for the contest (contestGetQualificationState returns "ready")
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a contest"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: group_id
//   in: path
//   type: integer
//   format: int64
//   required: true
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
		Duration                   string
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

		service.MustNotBeError(store.Exec(`
			INSERT INTO attempts (group_id, item_id, started_at, creator_id, `+"`order`"+`)
			SELECT ?, ?, ?, ?, IFNULL((SELECT MAX(`+"`order`"+`) FROM attempts WHERE group_id = ? AND item_id = ? FOR UPDATE), 0) + 1`,
			qualificationState.groupID, qualificationState.itemID,
			itemInfo.Now, srv.GetUser(r).GroupID, qualificationState.groupID, qualificationState.itemID).Error())

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
				VALUES(?, ?, DATE_ADD(?, INTERVAL (TIME_TO_SEC(?) + ?) SECOND))
				ON DUPLICATE KEY UPDATE expires_at = VALUES(expires_at)`,
				itemInfo.ContestParticipantsGroupID, qualificationState.groupID,
				itemInfo.Now, itemInfo.Duration, totalAdditionalTime).Error())
			service.MustNotBeError(store.GroupGroups().After())
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
