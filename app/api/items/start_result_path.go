package items

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation POST /items/{ids}/start-result-path items resultStartPath
//
//	---
//	summary: Start results for an item path
//	description: >
//		Creates new started results (or starts not started existing ones) for an item path if needed and returns the last attempt in the chain.
//
//		Of all possible chains of attempts the service chooses the one having missing/not-started results located closer
//		to the end of the path, preferring chains having less missing/not-started results and having higher values of `attempt_id`.
//		If there is no result for the first item, the service tries to create an attempt chain starting with the zero attempt.
//		The chain of attempts cannot have missing results for items requiring explicit entry or require to start/create results
//		within or below ended/not-allowing-submissions attempts.
//
//		If `as_team_id` is given, the created/updated results are linked to the `as_team_id` group instead of the user's self group.
//
//
//			Restrictions:
//
//		* if `as_team_id` is given, it should be a user's parent team group,
//		* the first item in `{ids}` should be a root activity/skill (groups.root_activity_id/root_skill_id) of a group
//			the participant is a descendant of or manages,
//		* `{ids}` should be an ordered list of parent-child items,
//		* the group starting results should have at least 'content' access on each of the items in `{ids}`,
//
//		otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: ids
//			in: path
//			type: string
//			description: slash-separated list of item IDs (no more than 15 IDs)
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"201":
//			description: "Created. Success response with the attempt id for the final item in the path"
//			schema:
//					type: object
//					required: [success, message, data]
//					properties:
//						success:
//							description: "true"
//							type: boolean
//							enum: [true]
//						message:
//							description: updated
//							type: string
//							enum: [updated]
//						data:
//							type: object
//							required: [attempt_id]
//							properties:
//								attempt_id:
//									description: The attempt linked to the final item in the path
//									type: integer
//									format: string
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
func (srv *Service) startResultPath(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var err error

	ids, err := idsFromRequest(httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	participantID := service.ParticipantIDFromContext(httpRequest.Context())

	var result []map[string]interface{}
	var attemptID int64
	var shouldSchedulePropagation bool
	store := srv.GetStore(httpRequest)
	err = store.InTransaction(func(store *database.DataStore) error {
		result = getDataForResultPathStart(store, participantID, ids)
		if len(result) == 0 {
			return service.ErrAPIInsufficientAccessRights // rollback
		}

		data := result[0]
		rowsToInsert := make([]map[string]interface{}, 0, len(data))
		rowsToInsertPropagate := make([]map[string]interface{}, 0, len(data))
		for index, itemID := range ids {
			attemptID = data[fmt.Sprintf("attempt_id%d", index)].(int64)       //nolint:forcetypeassert // we know it is int64
			if data[fmt.Sprintf("has_started_result%d", index)].(int64) == 1 { //nolint:forcetypeassert // we know it is int64
				continue
			}
			rowsToInsert = append(rowsToInsert, map[string]interface{}{
				"item_id":            itemID,
				"participant_id":     participantID,
				"attempt_id":         attemptID,
				"started_at":         database.Now(),
				"latest_activity_at": database.Now(),
			})
			rowsToInsertPropagate = append(rowsToInsertPropagate, map[string]interface{}{
				"item_id":        itemID,
				"participant_id": participantID,
				"attempt_id":     attemptID,
				"state":          "to_be_propagated",
			})
		}
		if len(rowsToInsert) > 0 {
			resultStore := store.Results()
			service.MustNotBeError(resultStore.InsertOrUpdateMaps(rowsToInsert, []string{"started_at", "latest_activity_at"}, nil))
			service.MustNotBeError(resultStore.DB.InsertMaps("results_propagate", rowsToInsertPropagate))
			shouldSchedulePropagation = true
		}

		return nil
	})
	service.MustNotBeError(err)

	if shouldSchedulePropagation {
		service.SchedulePropagation(store, srv.GetPropagationEndpoint(), []string{"results"})
	}

	service.MustNotBeError(render.Render(responseWriter, httpRequest, service.UpdateSuccess(map[string]interface{}{
		"attempt_id": strconv.FormatInt(attemptID, 10),
	})))
	return nil
}

func hasAccessToItemPath(store *database.DataStore, participantID int64, ids []int64) bool {
	var count int64
	store.Permissions().
		MatchingGroupAncestors(participantID).
		Joins(`JOIN items ON items.id = permissions.item_id`).
		Where("items.id IN (?)", ids).
		WherePermissionIsAtLeast("view", "content").
		Select("COUNT(DISTINCT items.id) AS count").
		Count(&count)

	return count == int64(len(ids))
}

func getDataForResultPathStart(store *database.DataStore, participantID int64, ids []int64) []map[string]interface{} {
	var result []map[string]interface{}
	if !hasAccessToItemPath(store, participantID, ids) {
		return result
	}

	participantAncestors := store.ActiveGroupAncestors().Where("child_group_id = ?", participantID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.ancestor_group_id").
		WithSharedWriteLock()
	groupsManagedByParticipant := store.ActiveGroupAncestors().ManagedByGroup(participantID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id").
		WithSharedWriteLock()
	rootActivities := participantAncestors.Select("groups.root_activity_id").Union(
		groupsManagedByParticipant.Select("groups.root_activity_id"))
	rootSkills := participantAncestors.Select("groups.root_skill_id").Union(
		groupsManagedByParticipant.Select("groups.root_skill_id"))

	query := store.Table("items as items0").WithSharedWriteLock()
	for i := 0; i < len(ids); i++ {
		query = query.Where(fmt.Sprintf("items%d.id = ?", i), ids[i])
	}
	query = query.Where("items0.id IN ? OR items0.id IN ?", rootActivities.SubQuery(), rootSkills.SubQuery())

	var score string
	var columns string
	attemptIsActiveCondition := "1"
	var columnsForOrder string
	for idIndex := 0; idIndex < len(ids); idIndex++ {
		var previousAttemptCondition string
		if idIndex > 0 {
			score += " + "
			const comma = ", "
			columns += comma
			// Chain link: when the attempt is rooted at this item the chain steps into a child attempt
			// (linked via parent_attempt_id); otherwise the same attempt id propagates from the previous rung.
			// The "false" branch also covers the relaxation below, where an explicit-entry item is matched on
			// a non-rooted attempt carrying a started result for it (e.g. attempt 0 propagating through the chain).
			previousAttemptCondition = fmt.Sprintf(` AND
					IF(attempts%[1]d.root_item_id = items%[1]d.id, attempts%[1]d.parent_attempt_id, attempts%[1]d.id) = attempts%[2]d.id`,
				idIndex, idIndex-1)
		}

		columnsForOrder += fmt.Sprintf(", attempts%d.id DESC", idIndex)
		attemptIsActiveCondition = fmt.Sprintf(
			"attempts%[1]d.ended_at IS NULL AND NOW() < attempts%[1]d.allows_submissions_until AND %[2]s",
			idIndex, attemptIsActiveCondition)
		score += fmt.Sprintf("((results%[1]d.started_at IS NULL) << %[2]d)", idIndex, len(ids)-idIndex-1)
		query = query.
			Joins(fmt.Sprintf("JOIN attempts AS attempts%[1]d ON attempts%[1]d.participant_id = ?"+previousAttemptCondition, idIndex),
				participantID).
			Joins(fmt.Sprintf(`
				LEFT JOIN results AS results%[1]d ON results%[1]d.participant_id = attempts%[1]d.participant_id AND
					attempts%[1]d.id = results%[1]d.attempt_id AND results%[1]d.item_id = items%[1]d.id`,
				idIndex)).
			// For items requiring explicit entry, the matched attempt usually must be rooted at the item itself
			// AND carry a result for it. We additionally allow non-rooted attempts when there is a STARTED result
			// for the item on the chosen attempt: such a started result is what proves the participant has actually
			// entered the item, even if the result is not on an attempt rooted at it (this can happen e.g. when
			// "requires_explicit_entry" was flipped on after the result was created). A non-started result on a
			// non-rooted attempt is intentionally NOT enough on its own. The two clauses below are kept separate
			// for clarity: the first picks which attempt is acceptable, the second enforces that the chosen
			// attempt actually has a result for the explicit-entry item.
			Where(fmt.Sprintf(
				"(NOT items%[1]d.requires_explicit_entry OR attempts%[1]d.root_item_id = items%[1]d.id "+
					"OR results%[1]d.started_at IS NOT NULL) AND "+
					"(NOT items%[1]d.requires_explicit_entry OR results%[1]d.attempt_id IS NOT NULL) AND "+
					"(results%[1]d.started_at IS NOT NULL OR %[2]s)",
				idIndex, attemptIsActiveCondition))

		if idIndex != len(ids)-1 {
			query = query.Joins(fmt.Sprintf(
				"JOIN items_items AS items_items%[2]d ON items_items%[2]d.parent_item_id = items%[1]d.id AND items_items%[2]d.child_item_id = ?",
				idIndex, idIndex+1), ids[idIndex+1]).
				Joins(fmt.Sprintf("JOIN items AS items%[1]d ON items%[1]d.id = items_items%[1]d.child_item_id", idIndex+1))
		}
		columns += fmt.Sprintf(
			"attempts%[1]d.id AS attempt_id%[1]d, results%[1]d.started_at IS NOT NULL AS has_started_result%[1]d", idIndex)
	}
	query = query.Select(columns).Where("results0.attempt_id IS NOT NULL OR attempts0.id = 0").
		Order(score + columnsForOrder).Limit(1)

	service.MustNotBeError(
		query.
			ScanIntoSliceOfMaps(&result).Error())

	return result
}
