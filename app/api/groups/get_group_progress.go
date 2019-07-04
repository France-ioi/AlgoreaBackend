package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/group-progress groups users attempts items getGroupProgress
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
//   items:
//     type: integer
// - name: from.name
//   description: Start the page from the group next to the group with `sName` = `from.name` and `ID` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the group next to the group with `sName`=`from.name` and `ID`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display results for the first N groups (sorted by `sName`)
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
	service.MustNotBeError(srv.Store.ItemItems().Where("idItemParent IN (?)", itemParentIDs).
		Joins("JOIN ? AS visible ON visible.idItem = items_items.idItemChild", itemsVisibleToUserSubQuery).
		Order("items_items.idItemChild").
		Pluck("items_items.idItemChild", &itemIDs).Error())
	if len(itemIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}
	itemsUnion := srv.Store.Raw("SELECT ? AS ID", itemIDs[0])
	for i := 1; i < len(itemIDs); i++ {
		itemsUnion = itemsUnion.UnionAll(srv.Store.Raw("SELECT ? AS ID", itemIDs[i]).QueryExpr())
	}

	// Preselect IDs of groups for that we will calculate the final stats.
	// All the "end members" are descendants of these groups.
	// There should not be too many of groups because we paginate on them.
	var ancestorGroupIDs []interface{}
	ancestorGroupIDQuery := srv.Store.GroupGroups().
		Where("groups_groups.idGroupParent = ?", groupID).
		Joins(`
			JOIN groups AS group_child
			ON group_child.ID = groups_groups.idGroupChild AND group_child.sType NOT IN('Team', 'UserSelf')`)
	ancestorGroupIDQuery, apiError := service.ApplySortingAndPaging(r, ancestorGroupIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "group_child.sName", FieldType: "string"},
		"id":   {ColumnName: "group_child.ID", FieldType: "int64"},
	}, "name,id")
	if apiError != service.NoError {
		return apiError
	}
	ancestorGroupIDQuery = service.NewQueryLimiter().
		SetDefaultLimit(10).SetMaxAllowedLimit(20).
		Apply(r, ancestorGroupIDQuery)
	service.MustNotBeError(ancestorGroupIDQuery.
		Pluck("group_child.ID", &ancestorGroupIDs).Error())
	if len(ancestorGroupIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	users := srv.Store.GroupGroups().
		Select("child.ID").
		Joins("JOIN groups AS parent ON parent.ID = groups_groups.idGroupParent AND parent.sType != 'Team'").
		Joins("JOIN groups AS child ON child.ID = groups_groups.idGroupChild and child.sType = 'UserSelf'").
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.idGroupAncestor IN (?) AND
				groups_ancestors.idGroupChild = parent.ID`, ancestorGroupIDs).
		Where("groups_groups.sType IN ('direct', 'invitationAccepted', 'requestAccepted')").
		Group("child.ID")

	teams := srv.Store.Table("groups FORCE INDEX(sType)").
		Select("groups.ID").
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.idGroupAncestor IN (?) AND
				groups_ancestors.idGroupChild = groups.ID`, ancestorGroupIDs).
		Where("groups.sType='Team'").
		Group("groups.ID")

	endMembers := users.Union(teams.SubQuery())

	endMembersStats := srv.Store.Raw(`
		SELECT
			end_members.ID,
			items.ID AS idItem,
			IFNULL(attempt_with_best_score.iScore, 0) AS iScore,
			IFNULL(attempt_with_best_score.bValidated, 0) AS bValidated,
			IFNULL(attempt_with_best_score.nbHintsCached, 0) AS nbHintsCached,
			IFNULL(attempt_with_best_score.nbSubmissionsAttempts, 0) AS nbSubmissionsAttempts,
			IF(attempt_with_best_score.idGroup IS NULL,
				0,
				(
					SELECT IF(attempt_with_best_score.bValidated,
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), MIN(sValidationDate)),
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), NOW())
					)
					FROM groups_attempts FORCE INDEX (GroupItemMinusScoreBestAnswerDateID)
					WHERE idGroup = end_members.ID AND idItem = items.ID
				)
			) AS iTimeSpent
		FROM ? AS end_members`, endMembers.SubQuery()).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.ID = (
				SELECT ID FROM groups_attempts FORCE INDEX (GroupItemMinusScoreBestAnswerDateID)
				WHERE idGroup = end_members.ID AND idItem = items.ID
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`)

	var dbResult []map[string]interface{}
	// It still takes more than 2 minutes to complete on large data sets
	service.MustNotBeError(
		srv.Store.GroupAncestors().
			Select(`
				groups_ancestors.idGroupAncestor AS idGroup,
				member_stats.idItem,
				AVG(member_stats.iScore) AS iAverageScore,
				AVG(member_stats.bValidated) AS iValidationRate,
				AVG(member_stats.nbHintsCached) AS iAvgHintsRequested,
				AVG(member_stats.nbSubmissionsAttempts) AS iAvgSubmissionsAttempts,
				AVG(member_stats.iTimeSpent) AS iAvgTimeSpent`).
			Joins("JOIN ? AS member_stats ON member_stats.ID = groups_ancestors.idGroupChild", endMembersStats.SubQuery()).
			Where("groups_ancestors.idGroupAncestor IN (?)", ancestorGroupIDs).
			Group("groups_ancestors.idGroupAncestor, member_stats.idItem").
			Order(gorm.Expr(
				"FIELD(groups_ancestors.idGroupAncestor"+strings.Repeat(", ?", len(ancestorGroupIDs))+")",
				ancestorGroupIDs...)).
			Order(gorm.Expr(
				"FIELD(member_stats.idItem"+strings.Repeat(", ?", len(itemIDs))+")",
				itemIDs...)).
			ScanIntoSliceOfMaps(&dbResult).Error())

	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
