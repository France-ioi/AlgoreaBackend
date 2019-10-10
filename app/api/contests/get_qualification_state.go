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
	// whether at least one member's ancestor group has now() in the `can_enter_from` -` can_enter_until` range for this item
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
	// whether at least one user's ancestor group has now() in the `can_enter_from` -` can_enter_until` range for this item
	// required: true
	CurrentUserCanEnter bool `json:"current_user_can_enter"`
	// required: true
	OtherMembers []contestGetQualificationStateOtherMember `json:"other_members"`

	groupID int64
	itemID  int64
}

// swagger:operation GET /contests/{item_id}/groups/{group_id}/qualification-state contests groups contestGetQualificationState
// ---
// summary: Get qualification state
// description: >
//                For the given contest and the given participant, returns the qualification state,
//                i.e. whether he can enter the contest, and info on each team member.
//
//                The qualification state is one of:
//                  * 'already_started' if the participant has a `contest_participations` row for the item;
//
//                  * 'not_ready' if there are more members than `contest_max_team_size` or
//                    if the team/user doesn't satisfy the contest entering condition which is computed
//                    in accordance with `items.contest_entering_condition` as follows:
//
//                      * "None": no additional conditions (the team/user can enter the contest);
//
//                      * "One": the current time needs to be included in the
//                        `groups_contest_items`.`can_enter_from`-`can_enter_until` time range
//                        for the contest item and one of the group ancestors of either the user (if participating alone)
//                        or at least one member of the team;
//
//                      * "All": same but all members of the team;
//
//                      * "Half": same but half of the members (ceil-rounded) of the team;
//
//                  * 'ready' otherwise.
//
//                Restrictions:
//                  * `item_id` should be a timed contest;
//                  * `group_id` should be either the current user's self group (if the item's `has_attempts` is false) or
//                     a team with `team_item_id` = `item_id` (otherwise);
//                  * the authenticated user should have at least grayed access to the item;
//                  * the authenticated user should be a member of the `group_id` (if it is a team).
//
//                Otherwise, the "Forbidden" response is returned.
// parameters:
// - name: item_id
//   description: "`id` of a timed contest"
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: group_id
//   in: path
//   type: integer
//   format: int64
//   required: true
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

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return nil, service.ErrInvalidRequest(err)
	}

	var contestInfo struct {
		IsTeamContest            bool `gorm:"column:has_attempts"`
		ContestMaxTeamSize       int32
		ContestEnteringCondition string
	}
	err = store.Items().VisibleByID(user, itemID).Where("items.duration IS NOT NULL").
		Select("items.has_attempts, items.contest_max_team_size, items.contest_entering_condition").
		Take(&contestInfo).Error()
	if gorm.IsRecordNotFoundError(err) {
		return nil, service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	if apiError := srv.checkGroupID(groupID, itemID, contestInfo.IsTeamContest, user, store); apiError != service.NoError {
		return nil, apiError
	}

	contestParticipationQuery := store.ContestParticipations().
		Where("item_id = ?", itemID).
		Where("group_id = ?", groupID)
	if lock {
		contestParticipationQuery = contestParticipationQuery.WithWriteLock()
	}
	alreadyStarted, err := contestParticipationQuery.HasRows()
	service.MustNotBeError(err)

	membersCount, otherMembers, currentUserCanEnter, qualifiedMembersCount :=
		srv.getQualificatonInfo(contestInfo.IsTeamContest, groupID, itemID, user, store)
	qualificationState := computeQualificationState(
		alreadyStarted, contestInfo.IsTeamContest, contestInfo.ContestMaxTeamSize,
		contestInfo.ContestEnteringCondition, membersCount, qualifiedMembersCount)

	result := &contestGetQualificationStateResponse{
		State:               qualificationState,
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

func (srv *Service) checkGroupID(
	groupID, itemID int64, isTeamContest bool, user *database.User, store *database.DataStore) service.APIError {
	if isTeamContest {
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
	} else if user.SelfGroupID == nil || groupID != *user.SelfGroupID {
		return service.InsufficientAccessRightsError
	}
	return service.NoError
}

func computeQualificationState(alreadyStarted, isTeamContest bool, maxTeamSize int32, contestEnteringCondition string,
	membersCount, qualifiedMembersCount int32) string {
	var qualificationState string
	if alreadyStarted {
		qualificationState = "already_started"
	} else {
		qualificationState = "ready"
		if isTeamContest && maxTeamSize < membersCount ||
			!isContestEnteringConditionSatisfied(contestEnteringCondition, membersCount, qualifiedMembersCount) {
			qualificationState = "not_ready"
		}
	}
	return qualificationState
}

func isContestEnteringConditionSatisfied(contestEnteringCondition string, membersCount, qualifiedMembersCount int32) bool {
	return contestEnteringCondition == "None" ||
		(contestEnteringCondition == "All" && qualifiedMembersCount == membersCount ||
			contestEnteringCondition == "Half" && membersCount <= qualifiedMembersCount*2 ||
			contestEnteringCondition == "One" && qualifiedMembersCount >= 1)
}

func (srv *Service) getQualificatonInfo(isTeamOnly bool, groupID, itemID int64, user *database.User, store *database.DataStore) (
	membersCount int32, otherMembers []contestGetQualificationStateOtherMember, currentUserCanEnter bool, qualifiedMembersCount int32) {
	if isTeamOnly {
		service.MustNotBeError(store.ActiveGroupGroups().Where("groups_groups_active.parent_group_id = ?", groupID).
			WhereActiveGroupRelationIsActual().
			Joins("JOIN users ON users.self_group_id = groups_groups_active.child_group_id").
			Joins(`
				LEFT JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.child_group_id`).
			Joins(`
					LEFT JOIN groups_contest_items ON groups_contest_items.group_id = groups_ancestors_active.ancestor_group_id AND
						groups_contest_items.item_id = ?`, itemID).
			Group("groups_groups_active.child_group_id").
			Order("groups_groups_active.child_group_id").
			Select(`
					users.first_name, users.last_name, users.self_group_id AS group_id, users.login,
					IFNULL(MAX(groups_contest_items.can_enter_from <= NOW() AND NOW() < groups_contest_items.can_enter_until), 0) AS can_enter`).
			Scan(&otherMembers).Error())
		membersCount = int32(len(otherMembers))
		var currentUserIndex int
		for index := range otherMembers {
			if otherMembers[index].GroupID == *user.SelfGroupID {
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
			Joins(`
					LEFT JOIN groups_contest_items ON groups_contest_items.group_id = groups_ancestors_active.ancestor_group_id
						AND groups_contest_items.item_id = ?`, itemID).
			Group("groups_ancestors_active.child_group_id").
			PluckFirst(`
					IFNULL(
						MAX(groups_contest_items.can_enter_from <= NOW() AND NOW() < groups_contest_items.can_enter_until), 0
					) AS can_enter`, &currentUserCanEnter).
			Error())
		if currentUserCanEnter {
			qualifiedMembersCount = 1
		}
	}
	return membersCount, otherMembers, currentUserCanEnter, qualifiedMembersCount
}
