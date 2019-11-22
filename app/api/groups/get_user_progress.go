package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupUserProgressResponseRow
type groupUserProgressResponseRow struct {
	// The user’s self `group_id`
	// required:true
	GroupID int64 `json:"group_id,string"`
	// required:true
	ItemID int64 `json:"item_id,string"`
	// The best score across all user's or user teams' attempts. If there are no attempts, the score is 0.
	// required:true
	Score float32 `json:"score"`
	// Whether the user or one of his teams has the item validated
	// required:true
	Validated bool `json:"validated"`
	// Nullable
	// required:true
	LatestActivityAt *database.Time `json:"latest_activity_at"`
	// Number of hints requested for the attempt with the best score (if multiple, take the first one, chronologically).
	// If there are no attempts, the number of hints is 0.
	// required:true
	HintsRequested int32 `json:"hints_requested"`
	// Number of submissions for the attempt with the best score (if multiple, take the first one, chronologically).
	// If there are no attempts, the number of submissions is 0.
	// required:true
	SubmissionsAttempts int32 `json:"submissions_attempts"`
	// Time spent by the user (or his teams) (in seconds):
	//
	//   1) if no attempts yet: 0
	//
	//   2) if one attempt validated: min(`validated_at`) - min(`started_at`)
	//     (i.e., time between the first time the user (or one of his teams) started one (any) attempt
	//      and the time he (or one of his teams) first validated the task)
	//
	//   3) if no attempts validated: `now` - min(`started_at`)
	// required:true
	TimeSpent int32 `json:"time_spent"`
}

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
//     description: OK. Success response with users progress on items
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupUserProgressResponseRow"
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

	itemsVisibleToUserSubQuery := srv.Store.Permissions().VisibleToUser(user).SubQuery()

	// Preselect IDs of end member for that we will calculate the stats.
	// There should not be too many of end members on one page.
	var userIDs []interface{}
	userIDQuery := srv.Store.ActiveGroupAncestors().
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id AND groups.type = 'UserSelf'").
		Where("groups_ancestors_active.ancestor_group_id = ?", groupID).
		Where("groups_ancestors_active.child_group_id != groups_ancestors_active.ancestor_group_id")
	userIDQuery, apiError := service.ApplySortingAndPaging(r, userIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "groups.name", FieldType: "string"},
		"id":   {ColumnName: "groups.id", FieldType: "int64"},
	}, "name,id", false)
	if apiError != service.NoError {
		return apiError
	}
	userIDQuery = service.NewQueryLimiter().Apply(r, userIDQuery)
	service.MustNotBeError(userIDQuery.
		Pluck("groups.id", &userIDs).Error())

	if len(userIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	// Preselect item IDs (there should not be many of them)
	var itemIDs []interface{}
	service.MustNotBeError(srv.Store.ItemItems().Where("parent_item_id IN (?)", itemParentIDs).
		Joins("JOIN ? AS visible ON visible.item_id = items_items.child_item_id", itemsVisibleToUserSubQuery).
		Order("items_items.child_item_id").
		Pluck("items_items.child_item_id", &itemIDs).Error())
	if len(itemIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}
	itemsUnion := srv.Store.Raw("SELECT ? AS id", itemIDs[0])
	for i := 1; i < len(itemIDs); i++ {
		itemsUnion = itemsUnion.UnionAll(srv.Store.Raw("SELECT ? AS id", itemIDs[i]).QueryExpr())
	}

	var result []groupUserProgressResponseRow
	service.MustNotBeError(srv.Store.Groups().
		Select(`
			items.id AS item_id,
			groups.id AS group_id,
			IFNULL(MAX(attempt_with_best_score.score), 0) AS score,
			IFNULL(MAX(attempt_with_best_score.validated), 0) AS validated,
			MAX(last_attempt.latest_activity_at) AS latest_activity_at,
			IFNULL(MAX(attempt_with_best_score.hints_cached), 0) AS hints_requested,
			IFNULL(MAX(attempt_with_best_score.submissions_attempts), 0) AS submissions_attempts,
			IF(MAX(attempt_with_best_score.group_id) IS NULL,
				0,
				IF(MAX(attempt_with_best_score.validated),
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.started_at), MIN(first_validated_attempt.validated_at)),
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.started_at), NOW())
				)
			) AS time_spent`).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_groups_active AS team_links
			ON team_links.child_group_id = groups.id`).
		Joins(`
			JOIN `+"`groups`"+` AS teams
			ON teams.type = 'Team' AND
				teams.id = team_links.parent_group_id`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, score, best_answer_at FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id
				ORDER BY group_id, item_id, score DESC, best_answer_at LIMIT 1
			) AS attempt_with_best_score_for_user ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, score, best_answer_at FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id
				ORDER BY group_id, item_id, score DESC, best_answer_at LIMIT 1
			) AS attempt_with_best_score_for_team ON 1`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.id = IF(attempt_with_best_score_for_team.score IS NOT NULL AND
				attempt_with_best_score_for_user.score IS NOT NULL AND (
				attempt_with_best_score_for_team.score > attempt_with_best_score_for_user.score OR
					(
						attempt_with_best_score_for_team.score = attempt_with_best_score_for_user.score AND
						attempt_with_best_score_for_team.best_answer_at < attempt_with_best_score_for_user.best_answer_at
					)
				) OR attempt_with_best_score_for_user.score IS NULL,
				attempt_with_best_score_for_team.id,
				attempt_with_best_score_for_user.id
			)`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, latest_activity_at FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND latest_activity_at IS NOT NULL
				ORDER BY latest_activity_at DESC LIMIT 1
			) AS last_attempt_of_user ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, latest_activity_at FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND latest_activity_at IS NOT NULL
				ORDER BY latest_activity_at DESC LIMIT 1
			) AS last_attempt_of_team ON 1`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt
			ON last_attempt.id = IF(
				(
					last_attempt_of_team.id IS NOT NULL AND
					last_attempt_of_user.id IS NOT NULL AND
					last_attempt_of_team.latest_activity_at > last_attempt_of_user.latest_activity_at
				) OR last_attempt_of_user.id IS NULL,
				last_attempt_of_team.id,
				last_attempt_of_user.id
			)`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, started_at FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND started_at IS NOT NULL
				ORDER BY started_at LIMIT 1
			) AS first_attempt_of_user ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, started_at FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND started_at IS NOT NULL
				ORDER BY started_at LIMIT 1
			) AS first_attempt_of_team ON 1`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt
			ON first_attempt.id = IF(
				(
					first_attempt_of_team.id IS NOT NULL AND
					first_attempt_of_user.id IS NOT NULL AND
					first_attempt_of_team.started_at < first_attempt_of_user.started_at
				) OR first_attempt_of_user.id IS NULL,
				first_attempt_of_team.id,
				first_attempt_of_user.id
			)`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, validated_at FROM groups_attempts
				WHERE group_id = groups.id AND item_id = items.id AND validated_at IS NOT NULL
				ORDER BY validated_at LIMIT 1
			) AS first_validated_attempt_of_user ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT id, validated_at FROM groups_attempts
				WHERE group_id = teams.id AND item_id = items.id AND validated_at IS NOT NULL
				ORDER BY validated_at LIMIT 1
			) AS first_validated_attempt_of_team ON 1`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt
			ON first_validated_attempt.id = IF(
				(
					first_validated_attempt_of_team.id IS NOT NULL AND
					first_validated_attempt_of_user.id IS NOT NULL AND
					first_validated_attempt_of_team.validated_at < first_validated_attempt_of_user.validated_at
				) OR first_attempt_of_user.id IS NULL,
				first_validated_attempt_of_team.id,
				first_validated_attempt_of_user.id
			)`).
		Where("groups.id IN (?)", userIDs).
		Group("groups.id, items.id").
		Order(gorm.Expr(
			"FIELD(groups.id"+strings.Repeat(", ?", len(userIDs))+")",
			userIDs...)).
		Order(gorm.Expr(
			"FIELD(items.id"+strings.Repeat(", ?", len(itemIDs))+")",
			itemIDs...)).
		Order("MIN(attempt_with_best_score.score), MAX(attempt_with_best_score.best_answer_at)").
		Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
