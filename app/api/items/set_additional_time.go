package items

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation PUT /items/{item_id}/groups/{group_id}/additional-times items itemSetAdditionalTime
//
//	---
//	summary: Set additional time for a time-limited item and a group
//	description: >
//							 For the input group and item, sets the `group_item_additional_times.additional_time` to the `time` value.
//							 If there is no `group_item_additional_times` for the given `group_id`, `item_id` and the `seconds` != 0, creates it
//							 (with default values in other columns).
//							 If no `group_item_additional_times` and `seconds` == 0, succeed without doing any change.
//
//
//							 `groups_groups.expires_at` & `attempts.allows_submissions_until` (for the latest attempt) of affected
//							 `items.participants_group_id` members is set to
//							 `results.started_at` + `items.duration` + total additional time.
//
//
//							 Restrictions:
//								 * `item_id` should be a time-limited item (with duration <> NULL);
//								 * the authenticated user should have `can_view` >= 'content' on the input item;
//								 * the authenticated user should have `can_grant_view` >= 'enter' on the input item;
//								 * the authenticated user should have `can_watch` >= 'result' on the input item;
//								 * the authenticated user should be a manager of the `group_id`
//									 with `can_grant_group_access` and `can_watch_members` permissions;
//								 * if the item is team-only (`items.entry_participant_type` = 'Team'), then the group should not be a user.
//
//							 Otherwise, the "Forbidden" response is returned.
//	parameters:
//		- name: item_id
//			description: "`id` of a time-limited item"
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: seconds
//			description: additional time in seconds (can be negative)
//			in: query
//			type: integer
//			minimum: -3020399
//			maximum: 3020399
//			required: true
//	responses:
//		"200":
//			"$ref": "#/responses/updatedResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) setAdditionalTime(w http.ResponseWriter, r *http.Request) error {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	itemID, groupID, seconds, err := srv.getParametersForSetAdditionalTime(r)
	service.MustNotBeError(err)

	var groupType string
	err = store.Groups().ManagedBy(user).Where("groups.id = ?", groupID).
		Having("MAX(can_grant_group_access) AND MAX(can_watch_members)").
		Group("groups.id").
		PluckFirst("groups.type", &groupType).Error()
	if gorm.IsRecordNotFoundError(err) {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var itemInfo struct {
		DurationInSeconds   int64
		IsTeamOnlyItem      bool
		ParticipantsGroupID int64
	}

	err = store.InTransaction(func(store *database.DataStore) error {
		err = store.Items().TimeLimitedByIDManagedByUser(itemID, user).WithExclusiveWriteLock().
			Select(`
				TIME_TO_SEC(items.duration) AS duration_in_seconds,
				items.entry_participant_type = 'Team' AS is_team_only_item,
				items.participants_group_id`).
			Take(&itemInfo).Error()
		if gorm.IsRecordNotFoundError(err) || (itemInfo.IsTeamOnlyItem && groupType == "User") {
			return service.InsufficientAccessRightsError // rollback
		}
		service.MustNotBeError(err)

		setAdditionalTimeForGroupAndTimeLimitedItem(store, groupID, itemID, itemInfo.ParticipantsGroupID,
			itemInfo.DurationInSeconds, seconds)
		return nil
	})
	service.MustNotBeError(err)

	render.Respond(w, r, service.UpdateSuccess[*struct{}](nil))
	return nil
}

func (srv *Service) getParametersForSetAdditionalTime(r *http.Request) (itemID, groupID, seconds int64, err error) {
	itemID, err = service.ResolveURLQueryPathInt64Field(r, "item_id")
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
	return itemID, groupID, seconds, nil
}

func setAdditionalTimeForGroupAndTimeLimitedItem(
	store *database.DataStore, groupID, itemID, participantsGroupID, durationInSeconds, additionalTimeInSeconds int64,
) {
	groupItemAdditionalTimeStore := store.GroupItemAdditionalTimes()
	scope := groupItemAdditionalTimeStore.Where("group_id = ?", groupID).Where("item_id = ?", itemID)
	found, err := scope.WithExclusiveWriteLock().HasRows()
	service.MustNotBeError(err)
	if found {
		service.MustNotBeError(scope.UpdateColumn("additional_time",
			gorm.Expr("SEC_TO_TIME(?)", additionalTimeInSeconds)).Error())
	} else if additionalTimeInSeconds != 0 {
		service.MustNotBeError(groupItemAdditionalTimeStore.Exec(
			"INSERT INTO group_item_additional_times (group_id, item_id, additional_time) VALUES(?, ?, SEC_TO_TIME(?))",
			groupID, itemID, additionalTimeInSeconds).Error())
	}

	service.MustNotBeError(store.Exec("DROP TEMPORARY TABLE IF EXISTS new_expires_at").Error())
	service.MustNotBeError(store.Exec(`
		CREATE TEMPORARY TABLE new_expires_at (
			child_group_id BIGINT(20) NOT NULL,
			expires_at DATETIME NOT NULL,
			PRIMARY KEY child_group_id (child_group_id)
		)`).Error())
	defer func() {
		// As we start from dropping the temporary table, it's optional to delete it here.
		// This means we can use a potentially canceled context and ignore the error.
		store.Exec("DROP TEMPORARY TABLE IF EXISTS new_expires_at")
	}()
	service.MustNotBeError(store.Exec(`
		INSERT INTO new_expires_at ?`,
		// For each of the groups participating/participated in solving the item ...
		store.GroupGroups().
			Where("groups_groups.parent_group_id = ?", participantsGroupID).
			// ... that are descendants of `groupID` (so affected by the change) ...
			Joins(`
				JOIN groups_ancestors_active AS changed_group_descendants
					ON changed_group_descendants.child_group_id = groups_groups.child_group_id AND
						changed_group_descendants.ancestor_group_id = ?`, groupID).
			// ... and have started solving the item ...
			Joins(`
				JOIN results
					ON results.participant_id = groups_groups.child_group_id AND
						results.started_at IS NOT NULL AND
						results.item_id = ?`, itemID).
			// ... and the attempt is not ended, ...
			Joins(`
				JOIN attempts
					ON attempts.participant_id = results.participant_id AND
					   attempts.id = results.attempt_id AND
					   attempts.ended_at IS NULL`).
			// ... we get all the ancestors to calculate the total additional time
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups.child_group_id").
			Joins(`
				JOIN group_item_additional_times
					ON group_item_additional_times.group_id = groups_ancestors_active.ancestor_group_id AND
						group_item_additional_times.item_id = results.item_id`).
			Group("groups_groups.child_group_id").
			Select(`
				groups_groups.child_group_id,
				DATE_ADD(
					MIN(results.started_at),
					INTERVAL (? + IFNULL(SUM(TIME_TO_SEC(group_item_additional_times.additional_time)), 0)) SECOND
				) AS expires_at`, durationInSeconds).
			WithExclusiveWriteLock().QueryExpr()).Error())

	// we always modify groups_groups.expires_at, no matter if it has been expired or not
	result := store.Exec(`
		UPDATE groups_groups
		JOIN new_expires_at
			ON new_expires_at.child_group_id = groups_groups.child_group_id
		SET groups_groups.expires_at = new_expires_at.expires_at
		WHERE groups_groups.parent_group_id = ?`, participantsGroupID)
	service.MustNotBeError(result.Error())

	groupsGroupsModified := result.RowsAffected() > 0

	// We are assuming here that a participant has only at most one ongoing participation at a moment.
	// This assumption impacts, for instance, this scenario:
	//
	//   * a user starts a first attempt at 2:00 which ends at 3:00,
	//   * the user starts a second attempt at 3:01 which will end at 4:01,
	//   * at 3:05, an admin adds 15min to the item -> only the second attempt gets the 15m extra.
	//
	// We only update attempts.allows_submission_until if the participation is active or
	// if the change makes it active.
	service.MustNotBeError(store.Exec(`
		UPDATE attempts
		JOIN new_expires_at
			ON new_expires_at.child_group_id = attempts.participant_id
		SET attempts.allows_submissions_until = new_expires_at.expires_at
		WHERE
			attempts.root_item_id = ? AND
			attempts.id =
				(SELECT id FROM (
					SELECT MAX(id) AS id FROM attempts WHERE participant_id = new_expires_at.child_group_id AND root_item_id = ? FOR UPDATE
				) AS latest_attempt) AND
			(NOW() < new_expires_at.expires_at OR NOW() < attempts.allows_submissions_until)
	`, itemID, itemID).Error())
	if groupsGroupsModified {
		service.MustNotBeError(store.GroupGroups().CreateNewAncestors())
		store.ScheduleResultsPropagation()
	}
}
