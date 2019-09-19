package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/user-progress users groups attempts items groupUserProgress
// ---
// summary: Display the current progress of users on a subset of items
// description: For all children of items from the parent_item_id list,
//              display the result of all user self-groups among the descendants of the given group
//              (including those in teams).
//
//              For each user, only the attempt corresponding to his best score counts
//              (across all his teams and his own attempts), disregarding whether or not
//              the score was done in a team which is descendant of the input group.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: parent_item_ids
//   required: true
//   in: query
//   type: array
//   items:
//     type: integer
// - name: from.name
//   description: Start the page from the user group next to the user group with `groups.name` = `from.name` and `groups.id` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the user group next to the user group with `groups.name`=`from.name` and `groups.id`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display results for the first N users (sorted by `groups.name`)
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     "$ref": "#/responses/groupsGetUserProgressResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getUserProgress(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemParentIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "parent_item_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemsVisibleToUserSubQuery := srv.Store.GroupItems().AccessRightsForItemsVisibleToUser(user).SubQuery()

	// Preselect IDs of end member for that we will calculate the stats.
	// There should not be too many of end members on one page.
	var userGroupIDs []interface{}
	userGroupIDQuery := srv.Store.GroupAncestors().
		Joins("JOIN `groups` ON groups.id = groups_ancestors.group_child_id AND groups.type = 'UserSelf'").
		Where("groups_ancestors.group_ancestor_id = ?", groupID).
		Where("groups_ancestors.group_child_id != groups_ancestors.group_ancestor_id")
	userGroupIDQuery, apiError := service.ApplySortingAndPaging(r, userGroupIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "groups.name", FieldType: "string"},
		"id":   {ColumnName: "groups.id", FieldType: "int64"},
	}, "name,id")
	if apiError != service.NoError {
		return apiError
	}
	userGroupIDQuery = service.NewQueryLimiter().Apply(r, userGroupIDQuery)
	service.MustNotBeError(userGroupIDQuery.
		Pluck("groups.id", &userGroupIDs).Error())

	if len(userGroupIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	// Preselect item IDs (there should not be many of them)
	var itemIDs []interface{}
	service.MustNotBeError(srv.Store.ItemItems().Where("item_parent_id IN (?)", itemParentIDs).
		Joins("JOIN ? AS visible ON visible.item_id = items_items.item_child_id", itemsVisibleToUserSubQuery).
		Order("items_items.item_child_id").
		Pluck("items_items.item_child_id", &itemIDs).Error())
	if len(itemIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}
	itemsUnion := srv.Store.Raw("SELECT ? AS id", itemIDs[0])
	for i := 1; i < len(itemIDs); i++ {
		itemsUnion = itemsUnion.UnionAll(srv.Store.Raw("SELECT ? AS id", itemIDs[i]).QueryExpr())
	}

	var dbResult []map[string]interface{}
	service.MustNotBeError(srv.Store.Groups().
		Select(`
			items.id AS item_id,
			groups.id AS group_id,
			IFNULL(MAX(attempt_with_best_score.score), 0) AS score,
			IFNULL(MAX(attempt_with_best_score.validated), 0) AS validated,
			MAX(last_attempt.last_activity_date) AS last_activity_date,
			IFNULL(MAX(attempt_with_best_score.hints_cached), 0) AS hints_requested,
			IFNULL(MAX(attempt_with_best_score.submissions_attempts), 0) AS submissions_attempts,
			IF(MAX(attempt_with_best_score.group_id) IS NULL,
				0,
				IF(MAX(attempt_with_best_score.validated),
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.start_date), MIN(first_validated_attempt.validation_date)),
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.start_date), NOW())
				)
			) AS time_spent`).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_groups AS team_links
			ON team_links.type`+database.GroupRelationIsActiveCondition+` AND
				team_links.group_child_id = groups.id`).
		Joins(`
			JOIN `+"`groups`"+` AS teams
			ON teams.type = 'Team' AND
				teams.id = team_links.group_parent_id`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score_for_user
			ON attempt_with_best_score_for_user.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id
				ORDER BY group_id, item_id, minus_score, best_answer_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score_for_team
			ON attempt_with_best_score_for_team.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id
				ORDER BY group_id, item_id, minus_score, best_answer_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.id = IF(attempt_with_best_score_for_team.score IS NOT NULL AND
				attempt_with_best_score_for_user.score IS NOT NULL AND (
				attempt_with_best_score_for_team.score > attempt_with_best_score_for_user.score OR
					(
						attempt_with_best_score_for_team.score = attempt_with_best_score_for_user.score AND
						attempt_with_best_score_for_team.best_answer_date < attempt_with_best_score_for_user.best_answer_date
					)
				) OR attempt_with_best_score_for_user.score IS NULL,
				attempt_with_best_score_for_team.id,
				attempt_with_best_score_for_user.id
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt_of_user
			ON last_attempt_of_user.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND last_activity_date IS NOT NULL
				ORDER BY last_activity_date DESC LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt_of_team
			ON last_attempt_of_team.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND last_activity_date IS NOT NULL
				ORDER BY last_activity_date DESC LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt
			ON last_attempt.id = IF(
				(
					last_attempt_of_team.id IS NOT NULL AND
					last_attempt_of_user.id IS NOT NULL AND
					last_attempt_of_team.last_activity_date > last_attempt_of_user.last_activity_date
				) OR last_attempt_of_user.id IS NULL,
				last_attempt_of_team.id,
				last_attempt_of_user.id
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt_of_user
			ON first_attempt_of_user.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND start_date IS NOT NULL
				ORDER BY start_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt_of_team
			ON first_attempt_of_team.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND start_date IS NOT NULL
				ORDER BY start_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt
			ON first_attempt.id = IF(
				(
					first_attempt_of_team.id IS NOT NULL AND
					first_attempt_of_user.id IS NOT NULL AND
					first_attempt_of_team.start_date < first_attempt_of_user.start_date
				) OR first_attempt_of_user.id IS NULL,
				first_attempt_of_team.id,
				first_attempt_of_user.id
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt_of_user
			ON first_validated_attempt_of_user.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND validation_date IS NOT NULL
				ORDER BY validation_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt_of_team
			ON first_validated_attempt_of_team.id = (
				SELECT id FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND validation_date IS NOT NULL
				ORDER BY validation_date LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt
			ON first_validated_attempt.id = IF(
				(
					first_validated_attempt_of_team.id IS NOT NULL AND
					first_validated_attempt_of_user.id IS NOT NULL AND
					first_validated_attempt_of_team.validation_date < first_validated_attempt_of_user.validation_date
				) OR first_attempt_of_user.id IS NULL,
				first_validated_attempt_of_team.id,
				first_validated_attempt_of_user.id
			)`).
		Where("groups.id IN (?)", userGroupIDs).
		Group("groups.id, items.id").
		Order(gorm.Expr(
			"FIELD(groups.id"+strings.Repeat(", ?", len(userGroupIDs))+")",
			userGroupIDs...)).
		Order(gorm.Expr(
			"FIELD(items.id"+strings.Repeat(", ?", len(itemIDs))+")",
			itemIDs...)).
		Order("MAX(attempt_with_best_score.minus_score), MAX(attempt_with_best_score.best_answer_date)").
		ScanIntoSliceOfMaps(&dbResult).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
