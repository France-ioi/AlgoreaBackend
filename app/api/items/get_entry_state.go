package items

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

type itemGetEntryStateOtherMember struct {
	*structures.UserPersonalInfo
	ShowPersonalInfo bool `json:"-"` // required: true

	Login string `json:"login"`
	// required: true
	GroupID int64 `json:"group_id,string"`
	// whether at least one member's ancestor group has now()
	// between `items.entering_time_min` (or `permissions_granted.can_enter_from`) and
	// `items.entering_time_max` (or `permissions_granted.can_enter_until`) range for this item
	// required: true
	CanEnter bool `json:"can_enter"`
	// true if the user has an active attempt as a member of another team for this item or
	// when the user has an expired attempt for this item as a member of another team
	// while the item doesn't allow multiple attempts
	AttemptsRestrictionViolated bool `json:"attempts_restriction_violated"`
}

// swagger:model itemGetEntryStateResponse
type itemGetEntryStateResponse struct {
	// required: true
	// enum: ready,already_started,not_ready
	State string `json:"state"`
	// `items.entry_max_team_size` (for team-only items)
	MaxTeamSize *int32 `json:"max_team_size,omitempty"`
	// required: true
	// enum: All,Half,One,None
	EntryMinAdmittedMembersRatio string `json:"entry_min_admitted_members_ratio"`
	// whether at least one user's ancestor group has NOW() between
	// `permissions_granted.can_enter_from` and `permissions_granted.can_enter_until`
	// and between `items.entering_time_min` and `items.entering_time_max` for this item
	// required: true
	CurrentUserCanEnter bool `json:"current_user_can_enter"`
	// required: true
	OtherMembers []itemGetEntryStateOtherMember `json:"other_members"`
	// whether a team should have frozen membership for entering the item (`items.entry_frozen_teams` = 1)
	// required: true
	FrozenTeamsRequired bool `json:"frozen_teams_required"`
	// whether the current team has frozen membership (`groups.frozen_membership` = 0)
	// required: true
	CurrentTeamIsFrozen bool `json:"current_team_is_frozen"`

	groupID int64
	itemID  int64
}

// swagger:operation GET /items/{item_id}/entry-state items itemGetEntryState
//
//	---
//	summary: Get entry state
//	description: >
//							 For the given item requiring explicit entry and the given participant
//							 (the current user or his team if `as_team_id` is set),
//							 returns the entry state, i.e. whether the participant can enter the item, and info on each team member.
//
//
//							 `first_name` and `last_name` of other members are only visible to managers of
//							 those members' groups to which they provided view access to personal data.
//
//
//							 The entry state is one of:
//								 * 'already_started' if the participant has an `attempts` row for the item
//									 (with `attempts.root_item_id` = `{item_id}`) allowing submissions;
//
//								 * 'not_ready' if the team itself cannot enter the item
//									 while there are more members than `entry_max_team_size` or
//									 the team/user doesn't satisfy the item's entry condition which is computed
//									 in accordance with `items.entry_min_admitted_members_ratio` as follows:
//
//										 * "None": no additional conditions (the team/user can enter the item);
//
//										 * "One": the current time needs to be between
//											 `permissions_granted.can_enter_from` and `permissions_granted.can_enter_until`
//											 and between `items.entering_time_min` and `items.entering_time_max`
//											 for the item and one of the group ancestors of either the user (if participating alone)
//											 or at least one member of the team;
//
//										 * "All": same but all members of the team;
//
//										 * "Half": same but half of the members (ceil-rounded) of the team;
//
//								 * 'not_ready' if the participant has an `attempts` row for the item (with `attempts.root_item_id` = `{item_id}`)
//									 while the item's `allows_multiple_attempts` is false;
//
//								 * 'not_ready' if at least one of the team's members as a member of another team
//									 has an `attempts` row for the item (with `attempts.root_item_id` = `{item_id}`)
//									 while the item's `allows_multiple_attempts` is false or an active (not expired) attempt;
//
//								 * 'not_ready' if the item's `items.entry_frozen_teams` = 1,
//									 but the team membership is not frozen (`groups.frozen_membership` = 0);
//
//								 * 'ready' otherwise.
//
//							 Restrictions:
//								 * `item_id` should require explicit entry;
//								 * `as_team_id` (if given) should be one of the current user's teams;
//								 * `as_team_id` should be given if the item is team-only and should not be given if the item is user-only;
//								 * the authenticated user (or his team) should have at least 'info' access to the item.
//
//							 Otherwise, the "Forbidden" response is returned.
//	parameters:
//		- name: item_id
//			description: "`id` of an item to enter"
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Success response with the entry state info
//			schema:
//				"$ref": "#/definitions/itemGetEntryStateResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getEntryState(w http.ResponseWriter, r *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	store := srv.GetStore(r)

	// We do not use the participant middleware as we get groups_groups.frozen_membership using the same SQL query
	groupID := user.GroupID
	if len(r.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "as_team_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
	}

	result, err := getItemInfoAndEntryState(itemID, groupID, user, store, false)
	service.MustNotBeError(err)

	render.Respond(w, r, result)
	return nil
}

func getItemInfoAndEntryState(itemID, groupID int64, user *database.User, store *database.DataStore, lock bool) (
	*itemGetEntryStateResponse, error,
) {
	var itemInfo struct {
		IsTeamItem                   bool
		AllowsMultipleAttempts       bool
		EntryMaxTeamSize             int32
		EntryMinAdmittedMembersRatio string
		EntryFrozenTeams             bool
	}

	err := store.Items().VisibleByID(groupID, itemID).Where("items.requires_explicit_entry").
		Select(`
			items.allows_multiple_attempts, items.entry_participant_type = 'Team' AS is_team_item,
			items.entry_max_team_size, items.entry_min_admitted_members_ratio, items.entry_frozen_teams`).
		Take(&itemInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return nil, service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	if (groupID != user.GroupID) != itemInfo.IsTeamItem {
		return nil, service.ErrAPIInsufficientAccessRights
	}

	var currentTeamHasFrozenMembership bool
	if groupID != user.GroupID {
		err = store.Groups().TeamGroupForUser(groupID, user).
			PluckFirst("frozen_membership", &currentTeamHasFrozenMembership).Error()
		if gorm.IsRecordNotFoundError(err) {
			return nil, service.ErrAPIInsufficientAccessRights
		}
		service.MustNotBeError(err)
	} else {
		itemInfo.EntryFrozenTeams = false // can be true only for team items
	}

	itemParticipationQuery := store.Attempts().
		Joins("JOIN items ON items.id = attempts.root_item_id").
		Where("attempts.root_item_id = ?", itemID).
		Where("attempts.participant_id = ?", groupID)
	if lock {
		itemParticipationQuery = itemParticipationQuery.WithExclusiveWriteLock()
	}
	var participationInfo struct {
		IsStarted bool
		IsActive  bool
	}
	err = itemParticipationQuery.Select(`
		IFNULL(MAX(1), 0) AS is_started,
		IFNULL(MAX(NOW() < attempts.allows_submissions_until), 0) AS is_active`).
		Scan(&participationInfo).Error()
	service.MustNotBeError(err)

	entryStateInfo := getEntryStateInfo(groupID, itemID, user, store, lock)
	state := computeEntryState(
		participationInfo.IsStarted, participationInfo.IsActive, itemInfo.AllowsMultipleAttempts, itemInfo.IsTeamItem,
		itemInfo.EntryMaxTeamSize, itemInfo.EntryMinAdmittedMembersRatio, entryStateInfo.membersCount, entryStateInfo.admittedMembersCount,
		entryStateInfo.attemptsViolationsFound, currentTeamHasFrozenMembership, itemInfo.EntryFrozenTeams, entryStateInfo.teamCanEnter)

	result := &itemGetEntryStateResponse{
		State:                        string(state),
		EntryMinAdmittedMembersRatio: itemInfo.EntryMinAdmittedMembersRatio,
		CurrentUserCanEnter:          entryStateInfo.currentUserCanEnter,
		OtherMembers:                 entryStateInfo.otherMembers,
		CurrentTeamIsFrozen:          currentTeamHasFrozenMembership,
		FrozenTeamsRequired:          itemInfo.EntryFrozenTeams,
		groupID:                      groupID,
		itemID:                       itemID,
	}
	if itemInfo.IsTeamItem {
		result.MaxTeamSize = &itemInfo.EntryMaxTeamSize
	}
	return result, nil
}

func computeEntryState(hasAlreadyStarted, isActive, allowsMultipleAttempts, isTeamItem bool,
	maxTeamSize int32, entryMinAdmittedMembersRatio string, membersCount, admittedMembersCount int32,
	attemptsViolationsFound, currentTeamIsFrozen, frozenTeamsRequired, teamCanEnter bool,
) entryState {
	if hasAlreadyStarted && isActive {
		return alreadyStarted
	}

	if isReadyToEnter(hasAlreadyStarted, allowsMultipleAttempts, isTeamItem,
		maxTeamSize, entryMinAdmittedMembersRatio, membersCount, admittedMembersCount,
		attemptsViolationsFound, currentTeamIsFrozen, frozenTeamsRequired, teamCanEnter) {
		return ready
	}

	return notReady
}

func isEntryMinAdmittedMembersRatioSatisfied(entryMinAdmittedMembersRatio string, membersCount, admittedMembersCount int32) bool {
	return entryMinAdmittedMembersRatio == "None" ||
		entryMinAdmittedMembersRatio == "All" && admittedMembersCount == membersCount ||
		entryMinAdmittedMembersRatio == "Half" && membersCount <= admittedMembersCount*2 ||
		entryMinAdmittedMembersRatio == "One" && admittedMembersCount >= 1
}

func isReadyToEnter(hasAlreadyStarted, allowsMultipleAttempts, isTeamItem bool,
	maxTeamSize int32, entryMinAdmittedMembersRatio string, membersCount, admittedMembersCount int32,
	attemptsViolationsFound, currentTeamIsFrozen, frozenTeamsRequired, teamCanEnter bool,
) bool {
	if isTeamItem && (!teamCanEnter && maxTeamSize < membersCount || frozenTeamsRequired && !currentTeamIsFrozen) ||
		!teamCanEnter && !isEntryMinAdmittedMembersRatioSatisfied(entryMinAdmittedMembersRatio, membersCount, admittedMembersCount) {
		return false
	}

	return !attemptsViolationsFound && (!hasAlreadyStarted || allowsMultipleAttempts)
}

type entryStateInfo struct {
	membersCount            int32
	otherMembers            []itemGetEntryStateOtherMember
	teamCanEnter            bool
	currentUserCanEnter     bool
	admittedMembersCount    int32
	attemptsViolationsFound bool
}

func getEntryStateInfo(groupID, itemID int64, user *database.User, store *database.DataStore, lock bool) *entryStateInfo {
	var result entryStateInfo
	if groupID != user.GroupID {
		result.teamCanEnter = discoverIfTeamCanEnter(groupID, itemID, store, lock)

		canEnterQuery := store.ActiveGroupGroups().Where("groups_groups_active.parent_group_id = ?", groupID).
			Joins("JOIN users ON users.group_id = groups_groups_active.child_group_id").
			Joins("JOIN items ON items.id = ?", itemID).
			Joins(`
				LEFT JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.child_group_id`).
			Joins(`
				LEFT JOIN permissions_granted ON permissions_granted.group_id = groups_ancestors_active.ancestor_group_id AND
					permissions_granted.item_id = items.id`).
			Group("groups_groups_active.child_group_id").
			Order("groups_groups_active.child_group_id").
			Select(`
				MAX(personal_info_view_approvals.approved) AS show_personal_info,
				IF(MAX(personal_info_view_approvals.approved), users.first_name, NULL) AS first_name,
				IF(MAX(personal_info_view_approvals.approved), users.last_name, NULL) AS last_name,
        users.group_id AS group_id, users.login,
				(? OR IFNULL(MAX(permissions_granted.can_enter_from <= NOW() AND NOW() < permissions_granted.can_enter_until), 0)) AND
				MAX(items.entering_time_min) <= NOW() AND NOW() < MAX(items.entering_time_max) AS can_enter`, result.teamCanEnter).
			WithPersonalInfoViewApprovals(user)
		if lock {
			canEnterQuery = canEnterQuery.WithExclusiveWriteLock()
		}
		service.MustNotBeError(canEnterQuery.Scan(&result.otherMembers).Error())
		result.membersCount = int32(len(result.otherMembers))

		participatingSomewhereElseQuery := store.ActiveGroupGroups().Where("groups_groups_active.parent_group_id = ?", groupID).
			Joins(`
				JOIN groups_groups_active AS all_teams_relations
				  ON all_teams_relations.child_group_id = groups_groups_active.child_group_id AND
				     all_teams_relations.is_team_membership = 1`).
			Joins("JOIN items ON items.id = ?", itemID).
			Joins("JOIN attempts ON attempts.participant_id = all_teams_relations.parent_group_id AND attempts.root_item_id = items.id").
			Where("all_teams_relations.parent_group_id != groups_groups_active.parent_group_id"). // except for this team
			Group("groups_groups_active.child_group_id").
			Having("MAX(NOW() < attempts.allows_submissions_until) OR NOT MAX(items.allows_multiple_attempts)")
		if lock {
			participatingSomewhereElseQuery = participatingSomewhereElseQuery.WithExclusiveWriteLock()
		}
		var usersViolatingAttemptsRestriction []int64
		service.MustNotBeError(participatingSomewhereElseQuery.
			Pluck("groups_groups_active.child_group_id", &usersViolatingAttemptsRestriction).Error())

		violationsMap := make(map[int64]bool, len(usersViolatingAttemptsRestriction))
		for _, userID := range usersViolatingAttemptsRestriction {
			violationsMap[userID] = true
		}

		result.attemptsViolationsFound = len(usersViolatingAttemptsRestriction) > 0

		var currentUserIndex int
		for index := range result.otherMembers {
			result.otherMembers[index].AttemptsRestrictionViolated = violationsMap[result.otherMembers[index].GroupID]
			if result.otherMembers[index].GroupID == user.GroupID {
				result.currentUserCanEnter = result.otherMembers[index].CanEnter
				currentUserIndex = index
			}
			if result.otherMembers[index].CanEnter {
				result.admittedMembersCount++
			}
			nilOtherMemberPersonalInfoIfNeeded(&result.otherMembers[index])
		}

		// remove the current user from the members list
		result.otherMembers = append(result.otherMembers[:currentUserIndex], result.otherMembers[currentUserIndex+1:]...)

		return &result
	}

	result.membersCount = 1
	result.otherMembers = []itemGetEntryStateOtherMember{}
	canEnterQuery := store.ActiveGroupAncestors().Where("groups_ancestors_active.child_group_id = ?", groupID).
		Joins("JOIN items ON items.id = ?", itemID).
		Joins(`
				LEFT JOIN permissions_granted ON permissions_granted.group_id = groups_ancestors_active.ancestor_group_id
					AND permissions_granted.item_id = items.id`).
		Group("groups_ancestors_active.child_group_id")
	if lock {
		canEnterQuery = canEnterQuery.WithExclusiveWriteLock()
	}

	service.MustNotBeError(canEnterQuery.
		PluckFirst(`
				IFNULL(
					MAX(permissions_granted.can_enter_from <= NOW() AND NOW() < permissions_granted.can_enter_until), 0
				) AND
				MAX(items.entering_time_min) <= NOW() AND NOW() < MAX(items.entering_time_max) AS can_enter`, &result.currentUserCanEnter).
		Error())
	if result.currentUserCanEnter {
		result.admittedMembersCount = 1
	}

	return &result
}

func nilOtherMemberPersonalInfoIfNeeded(otherMember *itemGetEntryStateOtherMember) {
	if !otherMember.ShowPersonalInfo {
		otherMember.UserPersonalInfo = nil
	}
}

func discoverIfTeamCanEnter(groupID, itemID int64, store *database.DataStore, lock bool) (teamCanEnter bool) {
	teamCanEnterQuery := store.ActiveGroupAncestors().Where("groups_ancestors_active.child_group_id = ?", groupID).
		Joins(`
				LEFT JOIN permissions_granted ON permissions_granted.group_id = groups_ancestors_active.ancestor_group_id AND
					permissions_granted.item_id = ?`, itemID).
		Joins("JOIN items ON items.id = ?", itemID)
	if lock {
		teamCanEnterQuery = teamCanEnterQuery.WithExclusiveWriteLock()
	}
	service.MustNotBeError(teamCanEnterQuery.PluckFirst(`
			IFNULL(MAX(permissions_granted.can_enter_from <= NOW() AND NOW() < permissions_granted.can_enter_until), 0) AND
			MAX(items.entering_time_min) <= NOW() AND NOW() < MAX(items.entering_time_max)`, &teamCanEnter).Error())
	return teamCanEnter
}
