package contests

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation PUT /contests/{item_id}/groups/{group_id}/additional-times contests contestSetAdditionalTime
// ---
// summary: Set additional time for a contest
// description: >
//                For the input group and item, sets the `groups_contest_items.additional_time` to the `time` value.
//                If there is no `groups_contest_items` for the given `group_id`, `item_id` and the `seconds` != 0, creates it
//                (with default values in other columns).
//                If no `groups_contest_items` and `seconds` == 0, succeed without doing any change.
//
//
//                `groups_groups.expires_at` of affected `items.contest_participants_group_id` members is set to
//                `attempts.entered_at` + `items.duration` + total additional time.
//
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * the authenticated user should have at least `content_with_descendants` access on the input item;
//                  * the authenticated user should own the `group_id`;
//                  * if the contest is team-only (`items.has_attempts` = 1), then the group should not be a user group.
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a timed contest"
//   in: path
//   type: integer
//   required: true
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: seconds
//   description: additional time in seconds (can be negative)
//   in: query
//   type: integer
//   minimum: -3020399
//   maximum: 3020399
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) setAdditionalTime(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, groupID, seconds, apiError := srv.getParametersForSetAdditionalTime(r)
	if apiError != service.NoError {
		return apiError
	}

	var groupType string
	err := srv.Store.Groups().ManagedBy(user).Where("groups.id = ?", groupID).
		PluckFirst("groups.type", &groupType).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var contestInfo struct {
		DurationInSeconds          int64
		IsTeamOnlyContest          bool
		ContestParticipantsGroupID int64
	}

	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		err = store.Items().ContestManagedByUser(itemID, user).WithWriteLock().
			Select(`
			TIME_TO_SEC(items.duration) AS duration_in_seconds,
			items.has_attempts AS is_team_only_contest,
			items.contest_participants_group_id`).
			Take(&contestInfo).Error()
		if gorm.IsRecordNotFoundError(err) || (contestInfo.IsTeamOnlyContest && groupType == "UserSelf") {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error
		}
		service.MustNotBeError(err)

		setAdditionalTimeForGroupInContest(store, groupID, itemID, contestInfo.ContestParticipantsGroupID,
			contestInfo.DurationInSeconds, seconds)
		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, service.UpdateSuccess(nil))
	return service.NoError
}

func (srv *Service) getParametersForSetAdditionalTime(r *http.Request) (itemID, groupID, seconds int64, apiError service.APIError) {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return 0, 0, 0, service.ErrInvalidRequest(err)
	}
	groupID, err = service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return 0, 0, 0, service.ErrInvalidRequest(err)
	}
	seconds, err = service.ResolveURLQueryGetInt64Field(r, "seconds")
	if err != nil {
		return 0, 0, 0, service.ErrInvalidRequest(err)
	}
	const maxSeconds = 838*3600 + 59*60 + 59
	// 838:59:59 is the maximum possible TIME value in MySQL
	if seconds < -maxSeconds || maxSeconds < seconds {
		return 0, 0, 0, service.ErrInvalidRequest(fmt.Errorf("'seconds' should be between %d and %d", -maxSeconds, maxSeconds))
	}
	return itemID, groupID, seconds, service.NoError
}

func setAdditionalTimeForGroupInContest(
	store *database.DataStore, groupID, itemID, participantsGroupID, durationInSeconds, additionalTimeInSeconds int64) {
	groupContestItemStore := store.GroupContestItems()
	scope := groupContestItemStore.Where("group_id = ?", groupID).Where("item_id = ?", itemID)
	found, err := scope.WithWriteLock().HasRows()
	service.MustNotBeError(err)
	if found {
		service.MustNotBeError(scope.UpdateColumn("additional_time",
			gorm.Expr("SEC_TO_TIME(?)", additionalTimeInSeconds)).Error())
	} else if additionalTimeInSeconds != 0 {
		service.MustNotBeError(groupContestItemStore.Exec(
			"INSERT INTO groups_contest_items (group_id, item_id, additional_time) VALUES(?, ?, SEC_TO_TIME(?))",
			groupID, itemID, additionalTimeInSeconds).Error())
	}

	service.MustNotBeError(store.Exec("DROP TEMPORARY TABLE IF EXISTS total_additional_times").Error())
	service.MustNotBeError(store.Exec(`
		CREATE TEMPORARY TABLE total_additional_times (
			PRIMARY KEY child_group_id (child_group_id)
		)
		?`,
		// For each of groups participating in the contest ...
		store.ActiveGroupGroups().
			Where("groups_groups_active.parent_group_id = ?", participantsGroupID).
			// ... that are descendants of `groupID` (so affected by the change) ...
			Joins(`
				JOIN groups_ancestors AS changed_group_descendants
					ON changed_group_descendants.child_group_id = groups_groups_active.child_group_id AND
						changed_group_descendants.ancestor_group_id = ?`, groupID).
			// ... and have entered the contest, ...
			Joins(`
				JOIN attempts AS contest_participations
					ON contest_participations.group_id = groups_groups_active.child_group_id AND
						contest_participations.entered_at IS NOT NULL AND
						contest_participations.item_id = ?`, itemID).
			// ... we get all the ancestors to calculate the total additional time
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.child_group_id").
			Joins(`
				JOIN groups_contest_items
					ON groups_contest_items.group_id = groups_ancestors_active.ancestor_group_id AND
						groups_contest_items.item_id = contest_participations.item_id`).
			Group("groups_groups_active.child_group_id").
			Select(`
				groups_groups_active.child_group_id,
				IFNULL(SUM(TIME_TO_SEC(groups_contest_items.additional_time)), 0) AS total_additional_time`).
			WithWriteLock().QueryExpr()).Error())

	//nolint:gosec
	result := store.Exec(`
		UPDATE groups_groups
		JOIN attempts AS contest_participations
			ON contest_participations.group_id = groups_groups.child_group_id AND
				contest_participations.item_id = ? AND contest_participations.entered_at IS NOT NULL AND
				contest_participations.entered_at IS NOT NULL
		JOIN total_additional_times
			ON total_additional_times.child_group_id = groups_groups.child_group_id
		SET groups_groups.expires_at = DATE_ADD(
			contest_participations.entered_at,
			INTERVAL (? + total_additional_times.total_additional_time) SECOND
		)
		WHERE NOW() < groups_groups.expires_at AND groups_groups.parent_group_id = ?`,
		itemID, durationInSeconds, participantsGroupID)
	service.MustNotBeError(result.Error())
	if result.RowsAffected() > 0 {
		service.MustNotBeError(store.GroupGroups().After())
	}
	service.MustNotBeError(store.Exec("DROP TEMPORARY TABLE total_additional_times").Error())
}
