package groups

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/user-progress users groups attempts items groupUserProgress
// ---
// summary: Display the current progress of users on a subset of items
// description: For all children of items from the parent_item_id list,
//              display the result of all user self-groups among the descendants of the given group
//              (including those in teams).
//
//              For all users, only the attempt corresponding to the best score counts
//              (across all his teams and his own attempts), disregarding whether or not
//              the score was done in a team which is descendant of the input group.
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
//   description: Start the page from the user group next to the user group with `groups.sName` = `from.name` and `groups.ID` = `from.id`
//                (`from.id` is required when `from.name` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the user group next to the user group with `groups.sName`=`from.name` and `groups.ID`=`from.id`
//                (`from.name` is required when from.id is present)
//   in: query
//   type: integer
// - name: limit
//   description: Display results for the first N users (sorted by `groups.sName`)
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
		Joins("JOIN groups ON groups.ID = groups_ancestors.idGroupChild AND groups.sType = 'UserSelf'").
		Where("groups_ancestors.idGroupAncestor = ?", groupID).
		Where("groups_ancestors.idGroupChild != groups_ancestors.idGroupAncestor")
	userGroupIDQuery, apiError := service.ApplySortingAndPaging(r, userGroupIDQuery, map[string]*service.FieldSortingParams{
		// Note that we require the 'from.name' request parameter although the service does not return group names
		"name": {ColumnName: "groups.sName", FieldType: "string"},
		"id":   {ColumnName: "groups.ID", FieldType: "int64"},
	}, "name,id")
	if apiError != service.NoError {
		return apiError
	}
	userGroupIDQuery = service.NewQueryLimiter().Apply(r, userGroupIDQuery)
	service.MustNotBeError(userGroupIDQuery.
		Pluck("groups.ID", &userGroupIDs).Error())

	if len(userGroupIDs) == 0 {
		render.Respond(w, r, []map[string]interface{}{})
		return service.NoError
	}

	var teamIDs []int64
	service.MustNotBeError(srv.Store.GroupGroups().
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.idGroupChild = groups_groups.idGroupParent AND
				groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild AND
				groups_ancestors.idGroupAncestor = ?`, groupID).
		Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent AND groups.sType = 'Team'").
		Where("groups_groups.idGroupChild IN (?)", userGroupIDs).
		Where("groups_groups.sType IN ('direct', 'requestAccepted', 'invitationAccepted')").
		Pluck("groups_groups.idGroupParent", &teamIDs).Error())

	// Preselect item IDs (there should not be many of them)
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

	var dbResult []map[string]interface{}
	service.MustNotBeError(srv.Store.Groups().
		Select(`
			items.ID AS idItem,
			groups.ID AS idGroup,
			IFNULL(attempt_with_best_score.iScore, 0) AS iScore,
			IFNULL(attempt_with_best_score.bValidated, 0) AS bValidated,
			MAX(last_attempt.sLastActivityDate) AS sLastActivityDate,
			IFNULL(attempt_with_best_score.nbHintsCached, 0) AS nbHintsRequested,
			IFNULL(attempt_with_best_score.nbSubmissionsAttempts, 0) AS nbSubmissionAttempts,
			IF(attempt_with_best_score.idGroup IS NULL,
				0,
				IF(attempt_with_best_score.bValidated,
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.sStartDate), MIN(first_validated_attempt.sValidationDate)),
					TIMESTAMPDIFF(SECOND, MIN(first_attempt.sStartDate), NOW())
				)
			) AS iTimeSpent`).
		Joins("JOIN ? AS items", itemsUnion.SubQuery()).
		Joins(`
			LEFT JOIN groups_groups AS team_links
			ON team_links.sType IN ('direct', 'requestAccepted', 'invitationAccepted') AND
				team_links.idGroupParent IN (?) AND
				team_links.idGroupChild = groups.ID
		`, teamIDs).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score_for_user
			ON attempt_with_best_score_for_user.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = groups.ID AND idItem = items.ID
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score_for_team
			ON attempt_with_best_score_for_team.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = team_links.idGroupParent AND idItem = items.ID
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.ID = IF(attempt_with_best_score_for_team.iScore IS NOT NULL AND
				attempt_with_best_score_for_user.iScore IS NOT NULL AND (
				attempt_with_best_score_for_team.iScore > attempt_with_best_score_for_user.iScore OR 
					(
						attempt_with_best_score_for_team.iScore = attempt_with_best_score_for_user.iScore AND
						attempt_with_best_score_for_team.sBestAnswerDate < attempt_with_best_score_for_user.sBestAnswerDate
					)
				) OR attempt_with_best_score_for_user.iScore IS NULL,
				attempt_with_best_score_for_team.ID,
				attempt_with_best_score_for_user.ID
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt_of_user
			ON last_attempt_of_user.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = groups.ID AND idItem = items.ID AND sLastActivityDate IS NOT NULL
				ORDER BY sLastActivityDate DESC LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt_of_team
			ON last_attempt_of_team.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = team_links.idGroupParent AND idItem = items.ID AND sLastActivityDate IS NOT NULL
				ORDER BY sLastActivityDate DESC LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS last_attempt
			ON last_attempt.ID = IF(
				(
					last_attempt_of_team.ID IS NOT NULL AND
					last_attempt_of_user.ID IS NOT NULL AND
					last_attempt_of_team.sLastActivityDate > last_attempt_of_user.sLastActivityDate
				) OR last_attempt_of_user.ID IS NULL,
				last_attempt_of_team.ID,
				last_attempt_of_user.ID
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt_of_user
			ON first_attempt_of_user.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = groups.ID AND idItem = items.ID AND sStartDate IS NOT NULL
				ORDER BY sStartDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt_of_team
			ON first_attempt_of_team.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = team_links.idGroupParent AND idItem = items.ID AND sStartDate IS NOT NULL
				ORDER BY sStartDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_attempt
			ON first_attempt.ID = IF(
				(
					first_attempt_of_team.ID IS NOT NULL AND
					first_attempt_of_user.ID IS NOT NULL AND
					first_attempt_of_team.sStartDate < first_attempt_of_user.sStartDate
				) OR first_attempt_of_user.ID IS NULL,
				first_attempt_of_team.ID,
				first_attempt_of_user.ID
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt_of_user
			ON first_validated_attempt_of_user.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = groups.ID AND idItem = items.ID AND sValidationDate IS NOT NULL
				ORDER BY sValidationDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt_of_team
			ON first_validated_attempt_of_team.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = team_links.idGroupParent AND idItem = items.ID AND sValidationDate IS NOT NULL
				ORDER BY sValidationDate LIMIT 1
			)`).
		Joins(`
			LEFT JOIN groups_attempts AS first_validated_attempt
			ON first_validated_attempt.ID = IF(
				(
					first_validated_attempt_of_team.ID IS NOT NULL AND
					first_validated_attempt_of_user.ID IS NOT NULL AND
					first_validated_attempt_of_team.sValidationDate < first_validated_attempt_of_user.sValidationDate
				) OR first_attempt_of_user.ID IS NULL,
				first_validated_attempt_of_team.ID,
				first_validated_attempt_of_user.ID
			)`).
		Where("groups.ID IN (?)", userGroupIDs).
		Group("groups.ID, items.ID").
		Order(gorm.Expr(
			"FIELD(groups.ID"+strings.Repeat(", ?", len(userGroupIDs))+")",
			userGroupIDs...)).
		Order(gorm.Expr(
			"FIELD(items.ID"+strings.Repeat(", ?", len(itemIDs))+")",
			itemIDs...)).
		Order("attempt_with_best_score.iMinusScore, attempt_with_best_score.sBestAnswerDate").
		ScanIntoSliceOfMaps(&dbResult).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
