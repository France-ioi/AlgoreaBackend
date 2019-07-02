package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

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

	var dbResult []map[string]interface{}
	// It still takes about 3 minutes to complete on large data sets
	service.MustNotBeError(srv.Store.GroupAncestors().
		Select(`
			groups_ancestors.idGroupAncestor AS idGroup,
			items.ID AS idItem,
			AVG(IFNULL(attempt_with_best_score.iScore, 0)) AS iAverageScore,
			AVG(IFNULL(attempt_with_best_score.bValidated, 0)) AS iValidationRate,
			AVG(IFNULL(attempt_with_best_score.nbHintsCached, 0)) AS iAvgHintsRequested,
			AVG(IFNULL(attempt_with_best_score.nbSubmissionsAttempts, 0)) AS iAvgSubmissionsAttempts,
			AVG(IF(attempt_with_best_score.idGroup IS NULL,
				0,
				(
					SELECT IF(MAX(bValidated),
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), MIN(sValidationDate)),
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), NOW())
					)
					FROM groups_attempts
					WHERE idGroup = end_member.ID AND idItem = items.ID
				)
			)) AS iAvgTimeSpent`).
		Joins(`
			JOIN groups AS end_member
			ON
				end_member.ID = groups_ancestors.idGroupChild AND
				end_member.sType IN ('UserSelf', 'Team')`).
		Joins(`
			JOIN (SELECT 1 as bKeepUser) AS keep_user
			ON end_member.sType = 'Team' OR (
				SELECT 1
				FROM groups_groups
				JOIN groups ON groups.ID = groups_groups.idGroupParent AND groups.sType != 'Team'
				WHERE
					idGroupChild = end_member.ID AND
					groups_groups.sType IN('invitationAccepted','requestAccepted','direct') AND
					idGroupParent IN (
							SELECT idGroupChild
							FROM groups_ancestors AS ga
							-- bIsSelf is good here since a user can be a direct member of the input group
							WHERE ga.idGroupAncestor = groups_ancestors.idGroupAncestor
						)
					LIMIT 1
				) = 1`).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = end_member.ID AND idItem = items.ID
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`).
		Where("groups_ancestors.idGroupChild != groups_ancestors.idGroupAncestor").
		Where("groups_ancestors.idGroupAncestor IN (?)", ancestorGroupIDs).
		Group("groups_ancestors.idGroupAncestor, items.ID").
		Order(gorm.Expr(
			"FIELD(groups_ancestors.idGroupAncestor"+strings.Repeat(", ?", len(ancestorGroupIDs))+")",
			ancestorGroupIDs...)).
		Order(gorm.Expr(
			"FIELD(items.ID"+strings.Repeat(", ?", len(itemIDs))+")",
			itemIDs...)).
		ScanIntoSliceOfMaps(&dbResult).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
