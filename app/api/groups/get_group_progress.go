package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getGroupProgress(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r) // hack it with: user := database.NewUser(1, srv.Store.Users(), nil)

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

	userSelfGroupID, err := user.SelfGroupID()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	var found bool
	found, err = srv.Store.Groups().OwnedBy(user).Where("groups.ID = ?", groupID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	itemsVisibleToUserSubquery := srv.Store.GroupItems().
		Select(
			"idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
				"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, "+
				"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess").
		Joins(`
			JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors
			ON ancestors.idGroupAncestor = groups_items.idGroup`, userSelfGroupID).
		Group("groups_items.idItem").
		Having("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0").SubQuery()

	// Preselect item IDs since we want to use them twice (for end members stats and for final stats)
	// There should not be many of them
	var itemIDs []int64
	service.MustNotBeError(srv.Store.ItemItems().Where("idItemParent IN (?)", itemParentIDs).
		Joins("JOIN ? AS visible ON visible.idItem = items_items.idItemChild", itemsVisibleToUserSubquery).
		Pluck("items_items.idItemChild", &itemIDs).Error())

	// Preselect IDs of groups for that we will calculate the final stats.
	// All the "end members" are descendants of these groups.
	// There should not be too many of groups because we paginate on them.
	var ancestorGroupIDs []int64
	ancestorGroupIDQuery := srv.Store.GroupGroups().
		Where("groups_groups.idGroupParent = ?", groupID).
		Joins(`
			JOIN groups AS group_child
			ON group_child.ID = groups_groups.idGroupChild AND group_child.sType NOT IN('Team', 'UserSelf')`)
	ancestorGroupIDQuery, apiError := service.ApplySortingAndPaging(r, ancestorGroupIDQuery, map[string]*service.FieldSortingParams{
		"id": {ColumnName: "group_child.ID", FieldType: "int64"},
	}, "id")
	if apiError != service.NoError {
		return apiError
	}
	ancestorGroupIDQuery = service.SetQueryLimit(r, ancestorGroupIDQuery, 10, 20)
	service.MustNotBeError(ancestorGroupIDQuery.
		Pluck("group_child.ID", &ancestorGroupIDs).Error())

	endMembersStats := srv.Store.
		Table("groups_attempts AS main_attempts FORCE INDEX(GroupItemMinusScoreBestAnswerDateID)").
		Select(`
			main_attempts.idItem AS idItem,
			main_attempts.idGroup AS idGroup,
			IFNULL(MAX(main_attempts.iScore), 0) AS iMaxEndMemberScore,
			IFNULL(MAX(main_attempts.bValidated), 0) AS bMaxEndMemberValidated,
			IFNULL(attempt_with_best_score.nbHintsCached, 0) AS nbEndMemberBestScoreHintsCached,
			IFNULL(attempt_with_best_score.nbSubmissionsAttempts, 0) AS nbEndMemberBestScoreSubmissionAttempts,
			IF(attempt_with_best_score.idGroup IS NULL,
				0,
				IF(MAX(main_attempts.bValidated),
					TIMESTAMPDIFF(SECOND, MIN(main_attempts.sStartDate), MIN(main_attempts.sValidationDate)),
					TIMESTAMPDIFF(SECOND, MIN(main_attempts.sStartDate), NOW())
				)
			) AS sEndMemberTimeSpent`).
		Joins(`
			JOIN groups_ancestors
			ON groups_ancestors.idGroupChild = main_attempts.idGroup AND
				groups_ancestors.idGroupChild != groups_ancestors.idGroupAncestor AND
				groups_ancestors.idGroupAncestor IN (?)`, ancestorGroupIDs). // preliminary filter (makes the query faster)
		Joins(`JOIN groups ON groups.ID = groups_ancestors.idGroupChild AND groups.sType IN ('UserSelf', 'Team')`).
		Joins(`
			JOIN groups_attempts AS attempt_with_best_score
			ON attempt_with_best_score.ID = (
				SELECT ID FROM groups_attempts
				WHERE idGroup = main_attempts.idGroup AND idItem = main_attempts.idItem
				ORDER BY idGroup, idItem, iMinusScore, sBestAnswerDate LIMIT 1
			)`).
		Where("main_attempts.idItem IN (?)", itemIDs). // preliminary filter (makes the query faster)
		Group("main_attempts.idGroup, main_attempts.idItem")

	endMembers := srv.Store.GroupAncestors().
		Select("groups_ancestors.idGroupAncestor AS idGroup, end_member.ID AS idEndMember").
		Joins(`
			JOIN groups AS end_member
			ON end_member.ID = groups_ancestors.idGroupChild AND end_member.sType IN ('UserSelf', 'Team')`).
		// The team users filtering starts here (is not well specified yet)
		Joins(`
			LEFT JOIN groups_groups AS team_link
			ON
				end_member.ID IS NOT NULL AND
				end_member.sType != 'Team' AND
				team_link.idGroupChild = end_member.ID AND
				team_link.sType IN('invitationAccepted','requestAccepted','direct')`).
		Joins(`
			LEFT JOIN groups_ancestors AS team_ancestor_check
			ON team_ancestor_check.idGroupChild = team_link.idGroupParent AND
				team_ancestor_check.idGroupAncestor = groups_ancestors.idGroupAncestor`).
		Joins("LEFT JOIN groups AS team ON team.ID = team_ancestor_check.idGroupChild AND team.sType = 'Team'").
		Joins(`
			LEFT JOIN (SELECT 1 as bHasOtherParents) AS other_parents
			ON team.ID IS NOT NULL AND (
				SELECT 1
				FROM groups_groups
				WHERE idGroupChild = end_member.ID AND
					idGroupParent != team.ID AND
					idGroupParent IN (SELECT idGroupChild FROM groups_ancestors AS ga WHERE ga.idGroupAncestor = groups_ancestors.idGroupAncestor)
				LIMIT 1
			) = 1`).
		// The team users filtering ends here
		Where("groups_ancestors.idGroupAncestor IN (?)", ancestorGroupIDs). // preliminary filter (makes the query faster)
		Where("groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild").
		Group("groups_ancestors.idGroupAncestor, end_member.ID").
		// The team users filtering (is not well specified yet)
		Having("MIN(end_member.sType = 'Team' OR (team.ID IS NULL OR (team.ID IS NOT NULL AND other_parents.bHasOtherParents)))")

	// We want to keep groups that don't have end members
	groupsEndMembers := srv.Store.Groups().
		Select("groups.ID AS idGroup, end_member.idEndMember AS idEndMember").
		Joins(`
			LEFT JOIN ? AS end_member
			ON end_member.idGroup = groups.ID`, endMembers.SubQuery()).
		Where("groups.ID IN (?)", ancestorGroupIDs)

	var dbResult []map[string]interface{}
	service.MustNotBeError(srv.Store.Raw(`
		SELECT
			groups_end_members.idGroup AS idGroup,
			items.ID AS idItem,
			AVG(IFNULL(end_members_stats.iMaxEndMemberScore, 0)) AS iAverageScore,
			AVG(IFNULL(end_members_stats.bMaxEndMemberValidated, 0)) AS iValidationRate,
			AVG(IFNULL(end_members_stats.nbEndMemberBestScoreHintsCached, 0)) AS iAvgHintsRequested,
			AVG(IFNULL(end_members_stats.nbEndMemberBestScoreSubmissionAttempts, 0)) AS iAvgSubmissionsAttempts,
			AVG(IFNULL(end_members_stats.sEndMemberTimeSpent, 0)) AS sAvgTimeSpent
		FROM ? AS groups_end_members
		JOIN items FORCE INDEX(PRIMARY) ON items.ID IN (?)
		LEFT JOIN ? AS end_members_stats
			ON end_members_stats.idGroup = groups_end_members.idEndMember AND end_members_stats.idItem = items.ID
		`, groupsEndMembers.SubQuery(), itemIDs, endMembersStats.SubQuery()).
		Group("groups_end_members.idGroup, items.ID").
		Order("groups_end_members.idGroup, items.ID").
		ScanIntoSliceOfMaps(&dbResult).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(dbResult)
	render.Respond(w, r, convertedResult)
	return service.NoError
}
