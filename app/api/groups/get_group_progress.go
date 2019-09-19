package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/group-progress groups users attempts items groupGroupProgress
// ---
// summary: Display the current progress of a group on a subset of items
// description: For all children of items from the parent_item_id list, display the result
//              of each direct child of the given `group_id` whose type is not in (Team,UserSelf).
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
//     "$ref": "#/responses/groupsGetGroupProgressResponse"
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

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	itemParentIDs, err := service.ResolveURLQueryGetInt64SliceField(r, "parent_item_ids")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	itemsVisibleToUserSubQuery := srv.Store.GroupItems().AccessRightsForItemsVisibleToUser(user).SubQuery()

	// Preselect item IDs since we want to use them twice (for end members stats and for final stats)
	// There should not be many of them
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

	// Preselect IDs of groups for that we will calculate the final stats.
	// All the "end members" are descendants of these groups.
	// There should not be too many of groups because we paginate on them.
	var ancestorGroupIDs []interface{}
	ancestorGroupIDQuery := srv.Store.GroupGroups().
		Where("groups_groups.group_parent_id = ?", groupID).
		Where("groups_groups.type = 'direct'").
		Joins(`
			JOIN ` + "`groups`" + ` AS group_child
			ON group_child.id = groups_groups.group_child_id AND group_child.type NOT IN('Team', 'UserSelf')`)
	ancestorGroupIDQuery, apiError := service.ApplySortingAndPaging(r, ancestorGroupIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "group_child.name", FieldType: "string"},
		"id":   {ColumnName: "group_child.id", FieldType: "int64"},
	}, "name,id")
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

	users := srv.Store.GroupGroups().
		Select("child.id").
		Joins("JOIN `groups` AS parent ON parent.id = groups_groups.group_parent_id AND parent.type != 'Team'").
		Joins("JOIN `groups` AS child ON child.id = groups_groups.group_child_id and child.type = 'UserSelf'").
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.group_ancestor_id IN (?) AND
				groups_ancestors.group_child_id = parent.id`, ancestorGroupIDs).
		WhereGroupRelationIsActive().
		Group("child.id")

	teams := srv.Store.Table("`groups` FORCE INDEX(type)").
		Select("groups.id").
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.group_ancestor_id IN (?) AND
				groups_ancestors.group_child_id = groups.id`, ancestorGroupIDs).
		Where("groups.type='Team'").
		Group("groups.id")

	endMembers := users.Union(teams.SubQuery())

	endMembersStats := srv.Store.Raw(`
		SELECT
			end_members.id,
			items.id AS item_id,
			IFNULL(attempt_with_best_score.score, 0) AS score,
			IFNULL(attempt_with_best_score.validated, 0) AS validated,
			IFNULL(attempt_with_best_score.hints_cached, 0) AS hints_cached,
			IFNULL(attempt_with_best_score.submissions_attempts, 0) AS submissions_attempts,
			IF(attempt_with_best_score.group_id IS NULL,
				0,
				(
					SELECT IF(attempt_with_best_score.validated,
						TIMESTAMPDIFF(SECOND, MIN(start_date), MIN(validation_date)),
						TIMESTAMPDIFF(SECOND, MIN(start_date), NOW())
					)
					FROM groups_attempts FORCE INDEX (group_item_minus_score_best_answer_date_id)
					WHERE group_id = end_members.id AND item_id = items.id
				)
			) AS time_spent
		FROM ? AS end_members`, endMembers.SubQuery()).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.id = (
				SELECT id FROM groups_attempts FORCE INDEX (group_item_minus_score_best_answer_date_id)
				WHERE group_id = end_members.id AND item_id = items.id
				ORDER BY group_id, item_id, minus_score, best_answer_date LIMIT 1
			)`)

	var dbResult []map[string]interface{}
	// It still takes more than 2 minutes to complete on large data sets
	service.MustNotBeError(
		srv.Store.GroupAncestors().
			Select(`
				groups_ancestors.group_ancestor_id AS group_id,
				member_stats.item_id,
				AVG(member_stats.score) AS average_score,
				AVG(member_stats.validated) AS validation_rate,
				AVG(member_stats.hints_cached) AS avg_hints_requested,
				AVG(member_stats.submissions_attempts) AS avg_submissions_attempts,
				AVG(member_stats.time_spent) AS avg_time_spent`).
			Joins("JOIN ? AS member_stats ON member_stats.id = groups_ancestors.group_child_id", endMembersStats.SubQuery()).
			Where("groups_ancestors.group_ancestor_id IN (?)", ancestorGroupIDs).
			Group("groups_ancestors.group_ancestor_id, member_stats.item_id").
			Order(gorm.Expr(
				"FIELD(groups_ancestors.group_ancestor_id"+strings.Repeat(", ?", len(ancestorGroupIDs))+")",
				ancestorGroupIDs...)).
			Order(gorm.Expr(
				"FIELD(member_stats.item_id"+strings.Repeat(", ?", len(itemIDs))+")",
				itemIDs...)).
			ScanIntoSliceOfMaps(&dbResult).Error())

	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
