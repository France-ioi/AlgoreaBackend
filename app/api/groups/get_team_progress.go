package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getTeamProgress(w http.ResponseWriter, r *http.Request) service.APIError {
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
	var teamIDs []interface{}
	teamIDQuery := srv.Store.GroupAncestors().
		Joins("JOIN groups ON groups.ID = groups_ancestors.idGroupChild AND groups.sType = 'Team'").
		Where("groups_ancestors.idGroupAncestor = ?", groupID).
		Where("groups_ancestors.idGroupChild != groups_ancestors.idGroupAncestor")
	teamIDQuery, apiError := service.ApplySortingAndPaging(r, teamIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "groups.sName", FieldType: "string"},
		"id":   {ColumnName: "groups.ID", FieldType: "int64"},
	}, "name,id")
	if apiError != service.NoError {
		return apiError
	}
	teamIDQuery = service.NewQueryLimiter().Apply(r, teamIDQuery)
	service.MustNotBeError(teamIDQuery.
		Pluck("groups.ID", &teamIDs).Error())

	if len(teamIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	itemsQuery := srv.Store.ItemItems().
		Select("items_items.idItemChild").
		Where("idItemParent IN (?)", itemParentIDs).
		Joins("JOIN ? AS visible ON visible.idItem = items_items.idItemChild", itemsVisibleToUserSubQuery)

	var dbResult []map[string]interface{}
	service.MustNotBeError(srv.Store.Groups().
		Select(`
			items.ID AS idItem,
			groups.ID AS idGroup,
			IFNULL(attempt_with_best_score.iScore, 0) AS iScore,
			IFNULL(attempt_with_best_score.bValidated, 0) AS bValidated,
			(SELECT MAX(sLastActivityDate) FROM groups_attempts WHERE idGroup = groups.ID AND idItem = items.ID) AS sLastActivityDate,
			IFNULL(attempt_with_best_score.nbHintsCached, 0) AS nbHintsRequested,
			IFNULL(attempt_with_best_score.nbSubmissionsAttempts, 0) AS nbSubmissionAttempts,
			IF(attempt_with_best_score.idGroup IS NULL,
				0,
				(
					SELECT IF(attempt_with_best_score.bValidated,
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), MIN(sValidationDate)),
						TIMESTAMPDIFF(SECOND, MIN(sStartDate), NOW())
					)
					FROM groups_attempts
					WHERE idGroup = groups.ID AND idItem = items.ID
				)
			) AS iTimeSpent`).
		Joins(`JOIN items ON items.ID IN ?`, itemsQuery.SubQuery()).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = groups.ID AND idItem = items.ID
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`).
		Where("groups.ID IN (?)", teamIDs).
		Order(gorm.Expr(
			"FIELD(groups.ID"+strings.Repeat(", ?", len(teamIDs))+")",
			teamIDs...)).
		Order("items.ID").
		ScanIntoSliceOfMaps(&dbResult).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
