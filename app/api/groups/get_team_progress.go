package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupTeamProgressResponseRow
type groupTeamProgressResponseRow struct {
	// The team’s `group_id`
	// required:true
	GroupID int64 `json:"group_id,string"`
	// required:true
	ItemID int64 `json:"item_id,string"`
	// Current score. If there are no results, the score is 0
	// required:true
	Score float32 `json:"score"`
	// Whether the team has the item validated
	// required:true
	Validated bool `json:"validated"`
	// Nullable
	// required:true
	LatestActivityAt *database.Time `json:"latest_activity_at"`
	// Number of hints requested for the result with the best score (if multiple, take the first one, chronologically).
	// If there are no results, the number of hints is 0.
	// required:true
	HintsRequested int32 `json:"hints_requested"`
	// Number of submissions for the result with the best score (if multiple, take the first one, chronologically).
	// If there are no results, the number of submissions is 0.
	// required:true
	Submissions int32 `json:"submissions"`
	// Time spent by the team (in seconds):
	//
	//   1) if no results yet: 0
	//
	//   2) if one result validated: min(`validated_at`) - min(`started_at`)
	//     (i.e., time between the first time it started one (any) attempt
	//      and the time he first validated the task)
	//
	//   3) if no results validated: `now` - min(`started_at`)
	// required:true
	TimeSpent int32 `json:"time_spent"`
}

// swagger:operation GET /groups/{group_id}/team-progress groups groupTeamProgress
// ---
// summary: Get group progress for teams
// description: >
//              Returns the current progress of teams on a subset of items.
//
//
//              For all visible children of items from the `{parent_item_id}` list,
//              displays the result of each team among the descendants of the group.
//
//
//              Restrictions:
//
//              * The current user should be a manager of the group (or of one of its ancestors)
//              with `can_watch_members` set to true,
//
//              * The current user should have `can_watch_members` >= 'result' on each of `{parent_item_ids}` items,
//
//
//              otherwise the 'forbidden' error is returned.
// parameters:
// - name: group_id
//   in: path
//   type: integer
//   required: true
// - name: parent_item_ids
//   in: query
//   required: true
//   type: array
//   items:
//     type: integer
// - name: from.name
//   description: Start the page from the team next to the team with `name` = `from.name` and `id` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the team next to the team with `name`=`from.name` and `id`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display results for the first N teams (sorted by `name`)
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with teams progress on items
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupTeamProgressResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getTeamProgress(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserCanWatchGroupMembers(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemParentIDs, apiError := srv.resolveAndCheckParentIDs(r, user)
	if apiError != service.NoError {
		return apiError
	}

	// Preselect IDs of end member for that we will calculate the stats.
	// There should not be too many of end members on one page.
	var teamIDs []interface{}
	teamIDQuery := srv.Store.ActiveGroupAncestors().
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.child_group_id AND groups.type = 'Team'").
		Where("groups_ancestors_active.ancestor_group_id = ?", groupID).
		Where("groups_ancestors_active.child_group_id != groups_ancestors_active.ancestor_group_id")
	teamIDQuery, apiError = service.ApplySortingAndPaging(r, teamIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "groups.name", FieldType: "string"},
		"id":   {ColumnName: "groups.id", FieldType: "int64"},
	}, "name,id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}
	teamIDQuery = service.NewQueryLimiter().Apply(r, teamIDQuery)
	service.MustNotBeError(teamIDQuery.
		Pluck("groups.id", &teamIDs).Error())

	if len(teamIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	itemsQuery := srv.Store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("view", "info").
		Joins("JOIN items_items ON items_items.child_item_id = permissions.item_id").
		Where("items_items.parent_item_id IN (?)", itemParentIDs).
		Select("DISTINCT items_items.child_item_id")

	var result []groupTeamProgressResponseRow
	service.MustNotBeError(srv.Store.Groups().
		Select(`
			items.id AS item_id,
			groups.id AS group_id,
			IFNULL(result_with_best_score.score_computed, 0) AS score,
			IFNULL(result_with_best_score.validated, 0) AS validated,
			(SELECT MAX(latest_activity_at) FROM results WHERE participant_id = groups.id AND item_id = items.id) AS latest_activity_at,
			IFNULL(result_with_best_score.hints_cached, 0) AS hints_requested,
			IFNULL(result_with_best_score.submissions, 0) AS submissions,
			IF(result_with_best_score.participant_id IS NULL,
				0,
				(
					SELECT GREATEST(IF(result_with_best_score.validated,
						TIMESTAMPDIFF(SECOND, MIN(started_at), MIN(validated_at)),
						TIMESTAMPDIFF(SECOND, MIN(started_at), NOW())
					), 0)
					FROM results
					WHERE participant_id = groups.id AND item_id = items.id
				)
			) AS time_spent`).
		Joins(`JOIN items ON items.id IN ?`, itemsQuery.SubQuery()).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT score_computed, validated, hints_cached, submissions, participant_id
				FROM results
				WHERE participant_id = groups.id AND item_id = items.id
				ORDER BY participant_id, item_id, score_computed DESC, score_obtained_at
				LIMIT 1
			) AS result_with_best_score ON 1`).
		Where("groups.id IN (?)", teamIDs).
		Order(gorm.Expr(
			"FIELD(groups.id"+strings.Repeat(", ?", len(teamIDs))+")",
			teamIDs...)).
		Order("items.id").
		Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}
