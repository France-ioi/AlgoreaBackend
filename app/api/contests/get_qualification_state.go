package contests

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type contestGetQualificationStateOtherMember struct {
	// Nullable
	// required: true
	FirstName *string `json:"first_name"`
	// Nullable
	// required: true
	LastName *string `json:"last_name"`
	// required: true
	Login string `json:"login"`
	// required: true
	GroupID int64 `json:"group_id,string"`
	// whether at least one member's ancestor group has now()
	// between `items.entering_time_min` (or `permissions_granted.can_enter_from`) and
	// `items.entering_time_max` (or `permissions_granted.can_enter_until`) range for this item
	// required: true
	CanEnter bool `json:"can_enter"`
}

// swagger:model contestGetQualificationStateResponse
type contestGetQualificationStateResponse struct {
	// required: true
	// enum: ready,already_started,not_ready
	State string `json:"state"`
	// `items.contest_max_team_size` (for team-only contests)
	MaxTeamSize *int32 `json:"max_team_size,omitempty"`
	// required: true
	// enum: All,Half,One,None
	EnteringCondition string `json:"entering_condition"`
	// whether at least one user's ancestor group has NOW() between
	// `permissions_granted.can_enter_from` and `permissions_granted.can_enter_until`
	// and between `items.entering_time_min` and `items.entering_time_max` for this item
	// required: true
	CurrentUserCanEnter bool `json:"current_user_can_enter"`
	// required: true
	OtherMembers []contestGetQualificationStateOtherMember `json:"other_members"`

	groupID int64
	itemID  int64
}

// swagger:operation GET /contests/{item_id}/qualification-state contests contestGetQualificationState
// ---
// summary: Get qualification state
// description: >
//                For the given contest and the given participant (the current user or his team if `as_team_id` is set),
//                returns the qualification state, i.e. whether the participant can enter the contest, and info on each team member.
//
//                The qualification state is one of:
//                  * 'already_started' if the participant has an `attempts` row for the item
//                    with non-null `started_at` and is an active member of the item's "contest participants" group;
//
//                  * 'not_ready' if there are more members than `contest_max_team_size` or
//                    if the team/user doesn't satisfy the contest entering condition which is computed
//                    in accordance with `items.contest_entering_condition` as follows:
//
//                      * "None": no additional conditions (the team/user can enter the contest);
//
//                      * "One": the current time needs to be between
//                        `permissions_granted.can_enter_from` and `permissions_granted.can_enter_until`
//                        and between `items.entering_time_min` and `items.entering_time_max`
//                        for the item and one of the group ancestors of either the user (if participating alone)
//                        or at least one member of the team;
//
//                      * "All": same but all members of the team;
//
//                      * "Half": same but half of the members (ceil-rounded) of the team;
//
//                  * 'not_ready' if the participant has an `attempts` row for the item
//                    with non-null `started_at` and is NOT an active member of the item's "contest participants" group
//                    while the item's `allows_multiple_attempts` is false;
//
//                  * 'ready' otherwise.
//
//                Restrictions:
//                  * `item_id` should be a contest;
//                  * `as_team_id` (if given) should be the current user's team having the `item_id` as the team item;
//                  * `as_team_id` should be given if the contest is team-only and should not be given if the contest is user-only;
//                  * the authenticated user (or his team) should have at least 'info' access to the item.
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a contest"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
//   format: int64
// responses:
//   "200":
//     description: OK. Success response with the qualification state info
//     schema:
//       "$ref": "#/definitions/contestGetQualificationStateResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getQualificationState(w http.ResponseWriter, r *http.Request) service.APIError {
	result, apiError := srv.getContestInfoAndQualificationStateFromRequest(r, srv.Store, false)
	if apiError != service.NoError {
		return apiError
	}

	render.Respond(w, r, result)
	return service.NoError
}

func (srv *Service) getContestInfoAndQualificationStateFromRequest(r *http.Request, store *database.DataStore, lock bool) (
	*contestGetQualificationStateResponse, service.APIError) {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return nil, service.ErrInvalidRequest(err)
	}

	groupID := user.GroupID
	if len(r.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(r, "as_team_id")
		if err != nil {
			return nil, service.ErrInvalidRequest(err)
		}
	}

	var contestInfo struct {
		IsTeamContest            bool
		AllowsMultipleAttempts   bool
		ContestMaxTeamSize       int32
		ContestEnteringCondition string
	}

	err = store.Items().VisibleByID(groupID, itemID).Where("items.requires_explicit_entry").
		Select(`
			items.allows_multiple_attempts, items.entry_participant_type = 'Team' AS is_team_contest,
			items.contest_max_team_size, items.contest_entering_condition`).
		Take(&contestInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return nil, service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if (groupID != user.GroupID) != contestInfo.IsTeamContest {
		return nil, service.InsufficientAccessRightsError
	}

	if apiError := srv.checkTeamID(groupID, itemID, user, store); apiError != service.NoError {
		return nil, apiError
	}

	contestParticipationQuery := store.Attempts().
		Joins("JOIN items ON items.id = attempts.item_id").
		// check the participation is not expired
		Joins(`
			LEFT JOIN groups_groups_active
				ON groups_groups_active.parent_group_id = items.contest_participants_group_id AND
					groups_groups_active.child_group_id = attempts.group_id`).
		Where("item_id = ?", itemID).
		Where("attempts.group_id = ?", groupID).
		Where("started_at IS NOT NULL")
	if lock {
		contestParticipationQuery = contestParticipationQuery.WithWriteLock()
	}
	var isActive, alreadyStarted bool
	err = contestParticipationQuery.PluckFirst("groups_groups_active.parent_group_id IS NOT NULL", &isActive).Error()
	if !gorm.IsRecordNotFoundError(err) {
		service.MustNotBeError(err)
		alreadyStarted = true
	}

	membersCount, otherMembers, currentUserCanEnter, qualifiedMembersCount :=
		srv.getQualificatonInfo(groupID, itemID, user, store)
	state := computeQualificationState(
		alreadyStarted, isActive, contestInfo.AllowsMultipleAttempts, contestInfo.IsTeamContest, contestInfo.ContestMaxTeamSize,
		contestInfo.ContestEnteringCondition, membersCount, qualifiedMembersCount)

	result := &contestGetQualificationStateResponse{
		State:               string(state),
		EnteringCondition:   contestInfo.ContestEnteringCondition,
		CurrentUserCanEnter: currentUserCanEnter,
		OtherMembers:        otherMembers,
		groupID:             groupID,
		itemID:              itemID,
	}
	if contestInfo.IsTeamContest {
		result.MaxTeamSize = &contestInfo.ContestMaxTeamSize
	}
	return result, service.NoError
}

func (srv *Service) checkTeamID(
	groupID, itemID int64, user *database.User, store *database.DataStore) service.APIError {
	if groupID != user.GroupID {
		var teamGroupID int64
		err := store.Groups().TeamGroupForTeamItemAndUser(itemID, user).
			PluckFirst("groups.id", &teamGroupID).Error()
		if gorm.IsRecordNotFoundError(err) {
			return service.InsufficientAccessRightsError
		}
		service.MustNotBeError(err)
		if teamGroupID != groupID {
			return service.InsufficientAccessRightsError
		}
	}
	return service.NoError
}

func computeQualificationState(hasAlreadyStarted, isActive, allowsMultipleAttempts, isTeamContest bool,
	maxTeamSize int32, contestEnteringCondition string, membersCount, qualifiedMembersCount int32) qualificationState {
	var state qualificationState
	if hasAlreadyStarted && isActive {
		state = alreadyStarted
	} else {
		state = ready
		if isTeamContest && maxTeamSize < membersCount ||
			!isContestEnteringConditionSatisfied(contestEnteringCondition, membersCount, qualifiedMembersCount) {
			state = notReady
		}
	}
	if hasAlreadyStarted && !isActive && !allowsMultipleAttempts {
		state = notReady
	}
	return state
}

func isContestEnteringConditionSatisfied(contestEnteringCondition string, membersCount, qualifiedMembersCount int32) bool {
	return contestEnteringCondition == "None" ||
		(contestEnteringCondition == "All" && qualifiedMembersCount == membersCount ||
			contestEnteringCondition == "Half" && membersCount <= qualifiedMembersCount*2 ||
			contestEnteringCondition == "One" && qualifiedMembersCount >= 1)
}

func (srv *Service) getQualificatonInfo(groupID, itemID int64, user *database.User, store *database.DataStore) (
	membersCount int32, otherMembers []contestGetQualificationStateOtherMember, currentUserCanEnter bool, qualifiedMembersCount int32) {
	if groupID != user.GroupID {
		service.MustNotBeError(store.ActiveGroupGroups().Where("groups_groups_active.parent_group_id = ?", groupID).
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
					users.first_name, users.last_name, users.group_id AS group_id, users.login,
					IFNULL(MAX(permissions_granted.can_enter_from <= NOW() AND NOW() < permissions_granted.can_enter_until), 0) AND
					MAX(items.entering_time_min) <= NOW() AND NOW() < MAX(items.entering_time_max) AS can_enter`).
			Scan(&otherMembers).Error())
		membersCount = int32(len(otherMembers))
		var currentUserIndex int
		for index := range otherMembers {
			if otherMembers[index].GroupID == user.GroupID {
				currentUserCanEnter = otherMembers[index].CanEnter
				currentUserIndex = index
			}
			if otherMembers[index].CanEnter {
				qualifiedMembersCount++
			}
		}
		// remove the current user from the members list
		otherMembers = append(otherMembers[:currentUserIndex], otherMembers[currentUserIndex+1:]...)
	} else {
		membersCount = 1
		otherMembers = []contestGetQualificationStateOtherMember{}
		service.MustNotBeError(store.ActiveGroupAncestors().Where("groups_ancestors_active.child_group_id = ?", groupID).
			Joins("JOIN items ON items.id = ?", itemID).
			Joins(`
					LEFT JOIN permissions_granted ON permissions_granted.group_id = groups_ancestors_active.ancestor_group_id
						AND permissions_granted.item_id = items.id`).
			Group("groups_ancestors_active.child_group_id").
			PluckFirst(`
					IFNULL(
						MAX(permissions_granted.can_enter_from <= NOW() AND NOW() < permissions_granted.can_enter_until), 0
					) AND
					MAX(items.entering_time_min) <= NOW() AND NOW() < MAX(items.entering_time_max) AS can_enter`, &currentUserCanEnter).
			Error())
		if currentUserCanEnter {
			qualifiedMembersCount = 1
		}
	}
	return membersCount, otherMembers, currentUserCanEnter, qualifiedMembersCount
}
