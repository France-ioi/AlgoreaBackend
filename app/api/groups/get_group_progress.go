package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model groupGroupProgressResponseRow
type groupGroupProgressResponseRow struct {
	// The child’s `group_id`
	// required:true
	GroupID int64 `json:"group_id,string"`
	// required:true
	ItemID int64 `json:"item_id,string"`
	// Average score of all "end-members".
	// The score of an "end-member" is the max of his `results.score` or 0 if no results.
	// required:true
	AverageScore float32 `json:"average_score"`
	// % (float [0,1]) of "end-members" who have validated the task.
	// An "end-member" has validated a task if one of his results has `results.validated` = 1.
	// No results for an "end-member" is considered as not validated.
	// required:true
	ValidationRate float32 `json:"validation_rate"`
	// Average number of hints requested by each "end-member".
	// The number of hints requested of an "end-member" is the `results.hints_cached`
	// of the result with the best score
	// (if several with the same score, we use the first result chronologically on `score_obtained_at`).
	// required:true
	AvgHintsRequested float32 `json:"avg_hints_requested"`
	// Average number of submissions made by each "end-member".
	// The number of submissions made by an "end-member" is the `results.submissions`.
	// of the result with the best score
	// (if several with the same score, we use the first result chronologically on `score_obtained_at`).
	// required:true
	AvgSubmissions float32 `json:"avg_submissions"`
	// Average time spent among all the "end-members" (in seconds). The time spent by an "end-member" is computed as:
	//
	//   1) if no results yet: 0
	//
	//   2) if one result validated: min(`validated_at`) - min(`started_at`)
	//     (i.e., time between the first time it started one (any) result
	//      and the time he first validated the task)
	//
	//   3) if no results validated: `now` - min(`started_at`)
	// required:true
	AvgTimeSpent float32 `json:"avg_time_spent"`
}

// swagger:operation GET /groups/{group_id}/group-progress groups groupGroupProgress
// ---
// summary: Get group progress
// description: >
//              Returns the current progress of a group on a subset of items.
//
//
//              For all visible children of items from the `{parent_item_id}` list, displays the result
//              of each direct child of the given `group_id` whose type is not in (Team, User).
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
//   type: array
//   required: true
//   items:
//     type: integer
// - name: from.name
//   description: Start the page from the group next to the group with `name` = `from.name` and `id` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the group next to the group with `name`=`from.name` and `id`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display results for the first N groups (sorted by `name`)
//   in: query
//   type: integer
//   maximum: 20
//   default: 10
// responses:
//   "200":
//     description: >
//       OK. Success response with groups progress on items
//       For all children of items in the parent_item_id list, display the result for each direct child
//       of the given group_id whose type is not in (Team, User). Values are averages of all the group's
//       "end-members" where “end-member” defined as descendants of the group which are either
//       1) teams or
//       2) users who descend from the input group not only through teams (one or more).
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupGroupProgressResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupProgress(w http.ResponseWriter, r *http.Request) service.APIError {
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

	// Preselect item IDs since we want to use them twice (for end members stats and for final stats)
	// There should not be many of them
	var itemIDs []interface{}
	service.MustNotBeError(srv.Store.Permissions().MatchingUserAncestors(user).
		Joins("JOIN items_items ON items_items.child_item_id = item_id").
		Where("items_items.parent_item_id IN (?)", itemParentIDs).
		WherePermissionIsAtLeast("view", "info").
		Order("items_items.child_item_id").
		Pluck("DISTINCT items_items.child_item_id", &itemIDs).Error())
	if len(itemIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}
	itemsUnion := srv.Store.Raw("SELECT ? AS id", itemIDs[0])
	for i := 1; i < len(itemIDs); i++ {
		itemsUnion = itemsUnion.UnionAll(srv.Store.Raw("SELECT ? AS id", itemIDs[i]).QueryExpr())
	}

	// Preselect IDs of groups for that we will calculate the final stats.
	// All the "end members" are descendants of these groups.
	// There should not be too many of groups because we paginate on them.
	var ancestorGroupIDs []interface{}
	ancestorGroupIDQuery := srv.Store.ActiveGroupGroups().
		Where("groups_groups_active.parent_group_id = ?", groupID).
		Joins(`
			JOIN ` + "`groups`" + ` AS group_child
			ON group_child.id = groups_groups_active.child_group_id AND group_child.type NOT IN('Team', 'User')`)
	ancestorGroupIDQuery, apiError = service.ApplySortingAndPaging(r, ancestorGroupIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "group_child.name", FieldType: "string"},
		"id":   {ColumnName: "group_child.id", FieldType: "int64"},
	}, "name,id", []string{"id"}, false)
	if apiError != service.NoError {
		return apiError
	}
	ancestorGroupIDQuery = service.NewQueryLimiter().
		SetDefaultLimit(10).SetMaxAllowedLimit(20).
		Apply(r, ancestorGroupIDQuery)
	service.MustNotBeError(ancestorGroupIDQuery.
		Pluck("group_child.id", &ancestorGroupIDs).Error())
	if len(ancestorGroupIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	endMembers := srv.Store.Groups().
		Select("groups.id").
		Joins(`
			JOIN groups_ancestors_active
			ON groups_ancestors_active.ancestor_group_id IN (?) AND
				groups_ancestors_active.child_group_id = groups.id`, ancestorGroupIDs).
		Where("groups.type = 'Team' OR groups.type = 'User'").
		Group("groups.id")

	endMembersStats := srv.Store.Raw(`
		SELECT
			end_members.id,
			items.id AS item_id,
			IFNULL(result_with_best_score.score, 0) AS score,
			IFNULL(result_with_best_score.validated, 0) AS validated,
			IFNULL(result_with_best_score.hints_cached, 0) AS hints_cached,
			IFNULL(result_with_best_score.submissions, 0) AS submissions,
			IF(result_with_best_score.participant_id IS NULL,
				0,
				(
					SELECT GREATEST(IF(result_with_best_score.validated,
						TIMESTAMPDIFF(SECOND, MIN(started_at), MIN(validated_at)),
						TIMESTAMPDIFF(SECOND, MIN(started_at), NOW())
					), 0)
					FROM results
					WHERE participant_id = end_members.id AND item_id = items.id
				)
			) AS time_spent
		FROM ? AS end_members`, endMembers.SubQuery()).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT score_computed AS score, validated, hints_cached, submissions, participant_id
				FROM results
				WHERE participant_id = end_members.id AND item_id = items.id
				ORDER BY participant_id, item_id, score_computed DESC, score_obtained_at
				LIMIT 1
			) AS result_with_best_score ON 1`)

	var result []groupGroupProgressResponseRow
	// It still takes more than 2 minutes to complete on large data sets
	service.MustNotBeError(
		srv.Store.ActiveGroupAncestors().
			Select(`
				groups_ancestors_active.ancestor_group_id AS group_id,
				member_stats.item_id,
				AVG(member_stats.score) AS average_score,
				AVG(member_stats.validated) AS validation_rate,
				AVG(member_stats.hints_cached) AS avg_hints_requested,
				AVG(member_stats.submissions) AS avg_submissions,
				AVG(member_stats.time_spent) AS avg_time_spent`).
			Joins("JOIN ? AS member_stats ON member_stats.id = groups_ancestors_active.child_group_id", endMembersStats.SubQuery()).
			Where("groups_ancestors_active.ancestor_group_id IN (?)", ancestorGroupIDs).
			Group("groups_ancestors_active.ancestor_group_id, member_stats.item_id").
			Order(gorm.Expr(
				"FIELD(groups_ancestors_active.ancestor_group_id"+strings.Repeat(", ?", len(ancestorGroupIDs))+")",
				ancestorGroupIDs...)).
			Order(gorm.Expr(
				"FIELD(member_stats.item_id"+strings.Repeat(", ?", len(itemIDs))+")",
				itemIDs...)).
			Scan(&result).Error())

	render.Respond(w, r, result)
	return service.NoError
}

func (srv *Service) resolveAndCheckParentIDs(r *http.Request, user *database.User) ([]int64, service.APIError) {
	itemParentIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "parent_item_ids")
	if err != nil {
		return nil, service.ErrInvalidRequest(err)
	}
	itemParentIDs = uniqueIDs(itemParentIDs)
	if len(itemParentIDs) > 0 {
		var cnt int
		service.MustNotBeError(srv.Store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("watch", "result").Where("item_id IN(?)", itemParentIDs).
			PluckFirst("COUNT(DISTINCT item_id)", &cnt).Error())
		if cnt != len(itemParentIDs) {
			return nil, service.InsufficientAccessRightsError
		}
	}
	return itemParentIDs, service.NoError
}

func uniqueIDs(ids []int64) []int64 {
	idsMap := make(map[int64]bool, len(ids))
	for _, id := range ids {
		idsMap[id] = true
	}
	ids = make([]int64, 0, len(idsMap))
	for id := range idsMap {
		ids = append(ids, id)
	}
	return ids
}
