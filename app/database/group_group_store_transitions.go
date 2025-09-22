package database

import (
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// GroupMembershipAction represents an action that changes relation between two groups.
type GroupMembershipAction string

const (
	// InvitationCreated means a pending group admin's invitation for user to join a group was created.
	InvitationCreated GroupMembershipAction = "invitation_created"
	// JoinRequestCreated means a pending user's request to join a group was created.
	JoinRequestCreated GroupMembershipAction = "join_request_created"
	// InvitationAccepted means a user became a member of a group by accepting an invitation.
	InvitationAccepted GroupMembershipAction = "invitation_accepted"
	// JoinRequestAccepted means a user became a member of a group since a group admin accepted his request.
	JoinRequestAccepted GroupMembershipAction = "join_request_accepted"
	// LeaveRequestAccepted means a user left a group since a group admin accepted his leave request.
	LeaveRequestAccepted GroupMembershipAction = "leave_request_accepted"
	// InvitationRefused means a user refused an invitation to join a group.
	InvitationRefused GroupMembershipAction = "invitation_refused"
	// InvitationWithdrawn means an admin withdrew his invitation to join a group.
	InvitationWithdrawn GroupMembershipAction = "invitation_withdrawn"
	// JoinedByBadge means a user has been added into a group because of a badge returned by the login module.
	JoinedByBadge GroupMembershipAction = "joined_by_badge"
	// JoinedByCode means a user joined a group by the group's code.
	JoinedByCode GroupMembershipAction = "joined_by_code"
	// JoinRequestRefused means an admin refused a user's request to join a group.
	JoinRequestRefused GroupMembershipAction = "join_request_refused"
	// JoinRequestWithdrawn means a user withdrew his request to join a group.
	JoinRequestWithdrawn GroupMembershipAction = "join_request_withdrawn"
	// Removed means a user was removed from a group.
	Removed GroupMembershipAction = "removed"
	// Left means a user left a group.
	Left GroupMembershipAction = "left"
	// IsMember means a user is a member of a group.
	IsMember GroupMembershipAction = "is_member"
	// LeaveRequestCreated means a pending user's request to leave a group was created.
	LeaveRequestCreated GroupMembershipAction = "is_member,leave_request_created"
	// LeaveRequestExpired means a pending user's leave request for an expired membership.
	LeaveRequestExpired GroupMembershipAction = "leave_request_created"
	// LeaveRequestRefused means a manager refused a user's request to leave a group.
	LeaveRequestRefused GroupMembershipAction = "leave_request_refused"
	// LeaveRequestWithdrawn means a user withdrew his request to leave a group.
	LeaveRequestWithdrawn GroupMembershipAction = "leave_request_withdrawn"
	// RemovedDueToApprovalChange means a user has been removed from a group because of approval changes.
	RemovedDueToApprovalChange GroupMembershipAction = "removed_due_to_approval_change"
	// NoRelation means there is no row for the group pair in the groups_groups/group_pending_requests tables.
	NoRelation GroupMembershipAction = ""
)

// PendingType converts the GroupMembershipAction into `group_pending_requests.type`.
func (groupMembershipAction GroupMembershipAction) PendingType() string {
	switch groupMembershipAction {
	case InvitationCreated:
		return "invitation"
	case JoinRequestCreated:
		return "join_request"
	case LeaveRequestCreated:
		return "leave_request"
	default:
		panic("groupMembershipAction should be of pending kind in PendingType()")
	}
}

func (groupMembershipAction GroupMembershipAction) isActive() bool {
	switch groupMembershipAction {
	case JoinedByBadge, InvitationAccepted, JoinRequestAccepted, JoinedByCode, IsMember,
		LeaveRequestCreated, LeaveRequestWithdrawn, LeaveRequestRefused:
		return true
	default:
		return false
	}
}

func (groupMembershipAction GroupMembershipAction) isPending() bool {
	switch groupMembershipAction {
	case InvitationCreated, JoinRequestCreated, LeaveRequestCreated, LeaveRequestExpired:
		return true
	default:
		return false
	}
}

func (groupMembershipAction GroupMembershipAction) hasApprovals() bool {
	return groupMembershipAction == JoinRequestCreated
}

// GroupGroupTransitionAction represents a groups_groups relation transition action.
type GroupGroupTransitionAction int

const (
	// AdminCreatesInvitation means a group admin invites new users to the group.
	AdminCreatesInvitation GroupGroupTransitionAction = iota
	// UserCreatesJoinRequest means a user creates a request to become a group member.
	UserCreatesJoinRequest
	// UserCreatesAcceptedJoinRequest means a user adds himself into a group that he owns
	// It doesn't check if the user owns the group / all needed approvals are given (a calling service should check that).
	UserCreatesAcceptedJoinRequest
	// UserAcceptsInvitation means a user accepts a group invitation.
	UserAcceptsInvitation
	// AdminAcceptsJoinRequest means a group admin accepts a request to join a group.
	// For this action we check that all the approvals required by the group are given in the join request
	// and set groups_groups.*_approved_at to group_pending_requests.at for each.
	AdminAcceptsJoinRequest
	// AdminAcceptsLeaveRequest means a group admin accepts a request to leave a group.
	AdminAcceptsLeaveRequest
	// AdminRefusesLeaveRequest means a group admin refuses a request to leave a group.
	AdminRefusesLeaveRequest
	// UserRefusesInvitation means a user refuses a group invitation.
	UserRefusesInvitation
	// AdminRefusesJoinRequest means a group admin refuses a request to join the group.
	AdminRefusesJoinRequest
	// AdminRemovesUser means a group admin removes a user from a group. It marks relations as "removed".
	// It doesn't check if a child is a user or not.
	AdminRemovesUser
	// AdminWithdrawsInvitation means a group admin withdraws an invitation.
	AdminWithdrawsInvitation
	// UserLeavesGroup means a user leaves a group.
	UserLeavesGroup
	// UserCreatesLeaveRequest means a user creates a request to leave a group
	// We don't check that groups.require_lock_membership_approval_until & groups_groups.lock_membership_approved_at
	// are not null (a calling service should check that by itself).
	UserCreatesLeaveRequest
	// UserCancelsJoinRequest means a user cancels his request to join a group.
	UserCancelsJoinRequest
	// UserCancelsLeaveRequest means a user cancels his request to leave a group.
	UserCancelsLeaveRequest
	// AdminRemovesDirectRelation removes a direct relation.
	AdminRemovesDirectRelation
	// UserJoinsGroupByBadge means we add a user into a group because of his badge returned by the login module.
	UserJoinsGroupByBadge
	// UserJoinsGroupByCode means a user joins a group using a group's code
	// We don't check the code here (a calling service should check the code by itself).
	UserJoinsGroupByCode
	// AdminStrengthensApprovalWithEmpty means an admin strengthens the approval requirements for a group and empties it.
	AdminStrengthensApprovalWithEmpty
	// AdminStrengthensApprovalWithReinvite means an admin strengthens the approval requirements for a group,
	// empties it and re-invites all the users.
	AdminStrengthensApprovalWithReinvite
)

type groupGroupTransitionRule struct {
	// Transitions defines all possible transitions for the action. The format is "FromAction->ToAction".
	// Relations that have "from" action not listed here are considered as invalid.
	Transitions map[GroupMembershipAction]GroupMembershipAction

	IfNotEnoughApprovalsDowngradeTo GroupMembershipAction
}

func groupGroupTransitionRuleForAction(action GroupGroupTransitionAction) groupGroupTransitionRule {
	return map[GroupGroupTransitionAction]groupGroupTransitionRule{
		AdminCreatesInvitation: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				NoRelation:          InvitationCreated,
				InvitationCreated:   InvitationCreated,
				JoinRequestCreated:  JoinRequestAccepted,
				LeaveRequestExpired: InvitationCreated,
			},
			IfNotEnoughApprovalsDowngradeTo: InvitationCreated, // only JoinRequestAccepted requires approvals.
		},
		UserCreatesJoinRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				NoRelation:          JoinRequestCreated,
				JoinRequestCreated:  JoinRequestCreated,
				LeaveRequestExpired: JoinRequestCreated,
			},
		},
		UserCreatesAcceptedJoinRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				NoRelation:          JoinRequestAccepted,
				JoinRequestCreated:  JoinRequestAccepted,
				InvitationCreated:   JoinRequestAccepted,
				LeaveRequestExpired: JoinRequestAccepted,
			},
		},
		UserJoinsGroupByBadge: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				NoRelation:          JoinedByBadge,
				JoinRequestCreated:  JoinedByBadge,
				InvitationCreated:   JoinedByBadge,
				LeaveRequestExpired: JoinedByBadge,
			},
		},
		UserJoinsGroupByCode: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				NoRelation:          JoinedByCode,
				JoinRequestCreated:  JoinedByCode,
				InvitationCreated:   JoinedByCode,
				LeaveRequestExpired: JoinedByCode,
			},
		},
		UserAcceptsInvitation: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				InvitationCreated: InvitationAccepted,
			},
		},
		AdminAcceptsJoinRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				JoinRequestCreated: JoinRequestAccepted,
			},
		},
		AdminAcceptsLeaveRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				LeaveRequestCreated: LeaveRequestAccepted,
			},
		},
		UserRefusesInvitation: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				InvitationCreated: InvitationRefused,
			},
		},
		AdminRefusesJoinRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				JoinRequestCreated: JoinRequestRefused,
			},
		},
		AdminRefusesLeaveRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				LeaveRequestCreated: LeaveRequestRefused,
			},
		},
		AdminRemovesUser: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            Removed,
				LeaveRequestCreated: Removed,
			},
		},
		AdminWithdrawsInvitation: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				InvitationCreated: InvitationWithdrawn,
			},
		},
		UserLeavesGroup: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            Left,
				LeaveRequestCreated: Left,
			},
		},
		UserCreatesLeaveRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            LeaveRequestCreated,
				LeaveRequestCreated: LeaveRequestCreated,
			},
		},
		UserCancelsJoinRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				JoinRequestCreated: JoinRequestWithdrawn,
			},
		},
		UserCancelsLeaveRequest: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				LeaveRequestCreated: LeaveRequestWithdrawn,
			},
		},
		AdminRemovesDirectRelation: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            NoRelation,
				NoRelation:          NoRelation,
				LeaveRequestCreated: NoRelation,
			},
		},
		AdminStrengthensApprovalWithEmpty: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            RemovedDueToApprovalChange,
				NoRelation:          NoRelation,
				JoinRequestCreated:  NoRelation,
				LeaveRequestCreated: RemovedDueToApprovalChange,
				LeaveRequestExpired: RemovedDueToApprovalChange,
				InvitationCreated:   NoRelation,
			},
		},
		AdminStrengthensApprovalWithReinvite: {
			Transitions: map[GroupMembershipAction]GroupMembershipAction{
				IsMember:            InvitationCreated,
				NoRelation:          NoRelation,
				JoinRequestCreated:  NoRelation,
				LeaveRequestCreated: RemovedDueToApprovalChange,
				LeaveRequestExpired: RemovedDueToApprovalChange,
				InvitationCreated:   InvitationCreated,
			},
		},
	}[action]
}

// GroupGroupTransitionResult is an enum{cycle, invalid, success, unchanged}.
type GroupGroupTransitionResult string

const (
	// Cycle means that the transition wasn't performed because it would create a cycle in groups_groups graph.
	Cycle GroupGroupTransitionResult = "cycle"
	// Invalid means that the transition is impossible.
	Invalid GroupGroupTransitionResult = "invalid"
	// ApprovalsMissing means that one or more approvals required by the transition are missing.
	ApprovalsMissing GroupGroupTransitionResult = "approvals_missing"
	// Full means that the parent group is full (in terms of `groups.max_participants`) when `enforce_max_participants` is true
	// (The number of participants is computed as the number of non-expired users or teams which are direct children
	//  of the group + invitations (join requests are not counted)).
	Full GroupGroupTransitionResult = "full"
	// Success means that the transition was performed successfully.
	Success GroupGroupTransitionResult = "success"
	// Unchanged means that the transition has been already performed.
	Unchanged GroupGroupTransitionResult = "unchanged"
)

// GroupGroupTransitionResults represents results of mass transition (format: map{ id -> GroupGroupTransitionResult }).
type GroupGroupTransitionResults map[int64]GroupGroupTransitionResult

// GroupApprovals represents all the approvals that can be given by a user to the group managers.
type GroupApprovals struct {
	PersonalInfoViewApproval bool
	LockMembershipApproval   bool
	WatchApproval            bool
}

// FromString initializes GroupApprovals from the given comma-separated list of approvals.
func (approvals *GroupApprovals) FromString(s string) {
	approvalsList := strings.Split(s, ",")
	for _, approval := range approvalsList {
		switch approval {
		case "personal_info_view":
			approvals.PersonalInfoViewApproval = true
		case "lock_membership":
			approvals.LockMembershipApproval = true
		case "watch":
			approvals.WatchApproval = true
		}
	}
}

// ToArray converts GroupApprovals to a list of approvals.
func (approvals *GroupApprovals) ToArray() []string {
	approvalsList := make([]string, 0, 3) //nolint:mnd // 3 possible approvals
	if approvals.PersonalInfoViewApproval {
		approvalsList = append(approvalsList, "personal_info_view")
	}
	if approvals.LockMembershipApproval {
		approvalsList = append(approvalsList, "lock_membership")
	}
	if approvals.WatchApproval {
		approvalsList = append(approvalsList, "watch")
	}
	return approvalsList
}

type stateInfo struct {
	ChildGroupID               int64
	Action                     GroupMembershipAction
	ApprovalsOK                bool
	PersonalInfoViewApprovedAt *Time
	LockMembershipApprovedAt   *Time
	WatchApprovedAt            *Time
}

type requiredApprovalsAndLimits struct {
	RequirePersonalInfoAccessApproval bool
	RequireLockMembershipApproval     bool
	RequireWatchApproval              bool
	EnforceMaxParticipants            bool
	MaxParticipants                   int
	IsTeam                            bool
}

// Transition performs a groups_groups relation transition according to groupGroupTransitionRules.
func (s *GroupGroupStore) Transition(action GroupGroupTransitionAction,
	parentGroupID int64, childGroupIDs []int64, approvals map[int64]GroupApprovals,
	performedByUserID int64,
) (results GroupGroupTransitionResults, approvalsToRequest map[int64]GroupApprovals, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	results = make(map[int64]GroupGroupTransitionResult, len(childGroupIDs))
	approvalsToRequest = make(map[int64]GroupApprovals, len(childGroupIDs))

	var oldStates []stateInfo
	var groupRequiredApprovalsAndLimits requiredApprovalsAndLimits
	dataStore := s.DataStore

	mustNotBeError(dataStore.Groups().ByID(parentGroupID).
		Select(`
				require_personal_info_access_approval != 'none' AS require_personal_info_access_approval,
				NOW() < IFNULL(require_lock_membership_approval_until, 0) AS require_lock_membership_approval,
				require_watch_approval, enforce_max_participants, max_participants, type = 'Team' AS is_team`).
		WithSharedWriteLock().Scan(&groupRequiredApprovalsAndLimits).Error())

	// Here we get current states for each childGroupID:
	// the current state can be one of
	// ("", "invitation_created", "join_request_created", "is_member", "is_member,leave_request_created")
	// where "is_member" means that childGroupID is a member of the parentGroupID
	mustNotBeError(
		dataStore.Raw(`
			SELECT child_group_id, GROUP_CONCAT(action) AS action,
				MAX(personal_info_view_approved_at) AS personal_info_view_approved_at,
				MAX(lock_membership_approved_at) AS lock_membership_approved_at, MAX(watch_approved_at) AS watch_approved_at
				FROM ((? FOR SHARE) UNION (? FOR UPDATE)) AS statuses
				GROUP BY child_group_id`,
			dataStore.ActiveGroupGroups().
				Select(`
					child_group_id, 'is_member' AS action,
					NULL AS personal_info_view_approved_at,
					NULL AS lock_membership_approved_at,
					NULL AS watch_approved_at`).
				Where("parent_group_id = ? AND child_group_id IN (?)", parentGroupID, childGroupIDs).QueryExpr(),
			dataStore.GroupPendingRequests().
				Select(`
					member_id,
					CASE group_pending_requests.type
						WHEN 'invitation' THEN 'invitation_created'
						WHEN 'join_request' THEN 'join_request_created'
						WHEN 'leave_request' THEN 'leave_request_created'
						ELSE group_pending_requests.type
					END,
					IF(personal_info_view_approved, at, NULL) AS personal_info_view_approved_at,
					IF(lock_membership_approved, at, NULL) AS lock_membership_approved_at,
					IF(watch_approved, at, NULL) AS watch_approved_at`).
				Where("group_id = ? AND member_id IN (?)", parentGroupID, childGroupIDs).QueryExpr()).
			Scan(&oldStates).Error())

	oldStatesMap := make(map[int64]stateInfo, len(childGroupIDs))
	for _, oldState := range oldStates {
		oldStatesMap[oldState.ChildGroupID] = oldState
	}

	// build the transition plan depending on the current states (oldStatesMap)
	plan := buildTransitionsPlan(
		parentGroupID, childGroupIDs, results, oldStatesMap, &groupRequiredApprovalsAndLimits, approvals, approvalsToRequest, action)

	performCyclesChecking(dataStore, parentGroupID, results, plan)

	enforceMaxSize(dataStore, action, parentGroupID, &groupRequiredApprovalsAndLimits, results, plan)

	shouldCreateNewAncestors := false
	shouldPropagatePermissions := false
	shouldPropagateResults := false

	deletePending(dataStore, plan.idsToDeletePending, parentGroupID)
	deleteRelations(dataStore, plan.idsToDeleteRelation, parentGroupID, &groupRequiredApprovalsAndLimits,
		&shouldCreateNewAncestors, &shouldPropagatePermissions)
	insertGroupPendingRequests(dataStore, plan.idsToInsertPending, parentGroupID, approvals)
	insertRelations(dataStore, plan.idsToInsertRelation, parentGroupID, approvals, oldStatesMap,
		&groupRequiredApprovalsAndLimits, &shouldCreateNewAncestors, &shouldPropagateResults)
	insertGroupMembershipChanges(dataStore, plan.idsChanged, parentGroupID, performedByUserID)

	if shouldCreateNewAncestors {
		dataStore.GroupGroups().createNewAncestors()
	}
	if shouldPropagatePermissions {
		dataStore.SchedulePermissionsPropagation()
	}
	if shouldPropagateResults {
		dataStore.ScheduleResultsPropagation()
	}

	return results, approvalsToRequest, nil
}

func deleteRelations(dataStore *DataStore, idsToDeleteRelation *golang.Set[int64], parentGroupID int64,
	groupRequiredApprovalsAndLimits *requiredApprovalsAndLimits, shouldCreateNewAncestors, shouldPropagatePermissions *bool,
) {
	if !idsToDeleteRelation.IsEmpty() {
		idsToDeleteSlice := idsToDeleteRelation.Values()
		mustNotBeError(dataStore.GroupGroups().Delete("parent_group_id = ? AND child_group_id IN (?)", parentGroupID, idsToDeleteSlice).Error())
		result := dataStore.PermissionsGranted().
			Delete("origin = 'group_membership' AND source_group_id = ? AND group_id IN (?)", parentGroupID, idsToDeleteSlice)
		mustNotBeError(result.Error())
		if !groupRequiredApprovalsAndLimits.IsTeam {
			*shouldCreateNewAncestors = true
		}
		*shouldPropagatePermissions = result.RowsAffected() > 0
	}
}

func deletePending(dataStore *DataStore, idsToDeletePending *golang.Set[int64], parentGroupID int64) {
	if !idsToDeletePending.IsEmpty() {
		idsToDeleteSlice := idsToDeletePending.Values()
		mustNotBeError(dataStore.GroupPendingRequests().Delete("group_id = ? AND member_id IN (?)", parentGroupID, idsToDeleteSlice).Error())
	}
}

func insertRelations(dataStore *DataStore, idsToInsertRelation *golang.Set[int64], parentGroupID int64,
	approvals map[int64]GroupApprovals, oldStatesMap map[int64]stateInfo, groupRequiredApprovalsAndLimits *requiredApprovalsAndLimits,
	shouldCreateNewAncestors, shouldPropagateResults *bool,
) {
	if !idsToInsertRelation.IsEmpty() {
		idsToInsertRelation := idsToInsertRelation.Values()
		insertQuery := `
			INSERT INTO groups_groups (
				parent_group_id, child_group_id, personal_info_view_approved_at,
				lock_membership_approved_at, watch_approved_at
			)`
		valuesTemplate := `(?, ?, ?, ?, ?)`
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsToInsertRelation)-1) +
			valuesTemplate // #nosec
		insertQuery += " ON DUPLICATE KEY UPDATE expires_at = '9999-12-31 23:59:59'"
		values := make([]interface{}, 0, len(idsToInsertRelation)*5) //nolint:mnd // 5 values per row
		for _, id := range idsToInsertRelation {
			personalInfoViewApprovedAt, lockMembershipApprovedAt, watchApprovedAt := resolveApprovalTimesForGroupsGroups(
				oldStatesMap, id, approvals,
			)
			values = append(values, parentGroupID, id,
				personalInfoViewApprovedAt, lockMembershipApprovedAt, watchApprovedAt)
		}
		if !groupRequiredApprovalsAndLimits.IsTeam {
			*shouldCreateNewAncestors = true
			*shouldPropagateResults = true
		}
		mustNotBeError(dataStore.Exec(insertQuery, values...).Error())
	}
}

func enforceMaxSize(dataStore *DataStore, action GroupGroupTransitionAction, parentGroupID int64,
	limits *requiredApprovalsAndLimits, results GroupGroupTransitionResults, plan *transitionsPlan,
) {
	if !limits.EnforceMaxParticipants || !map[GroupGroupTransitionAction]bool{
		UserJoinsGroupByBadge: true, UserJoinsGroupByCode: true, UserCreatesJoinRequest: true, UserCreatesAcceptedJoinRequest: true,
		AdminCreatesInvitation: true, AdminAcceptsJoinRequest: true,
	}[action] {
		return
	}

	changedIDsList := make([]int64, 0, len(plan.idsChanged))
	for id := range plan.idsChanged {
		changedIDsList = append(changedIDsList, id)
	}
	var activeRelationsCount int
	mustNotBeError(dataStore.ActiveGroupGroups().Where("parent_group_id = ?", parentGroupID).
		Where("child_group_type IN ('User', 'Team')").
		Where("child_group_id NOT IN(?)", changedIDsList).WithSharedWriteLock().Count(&activeRelationsCount).Error())
	var invitationsCount int
	mustNotBeError(dataStore.GroupPendingRequests().
		Where("group_id = ?", parentGroupID).
		Where("type = 'invitation'").Where("member_id NOT IN(?)", changedIDsList).
		WithSharedWriteLock().Count(&invitationsCount).Error())

	membersCount := activeRelationsCount + invitationsCount
	for _, itemAction := range plan.idsChanged {
		if itemAction.isActive() || itemAction == InvitationCreated {
			membersCount++
		}
	}
	if membersCount > limits.MaxParticipants || membersCount == limits.MaxParticipants && action == UserCreatesJoinRequest {
		deleteIDsFromTransitionPlan(changedIDsList, Full, results, plan)
	}
}

func resolveApprovalTimesForGroupsGroups(oldStatesMap map[int64]stateInfo, groupID int64, approvals map[int64]GroupApprovals) (
	personalInfoViewApprovedAt, lockMembershipApprovedAt, watchApprovedAt interface{},
) {
	personalInfoViewApprovedAt = gorm.Expr("NULL")
	lockMembershipApprovedAt = gorm.Expr("NULL")
	watchApprovedAt = gorm.Expr("NULL")
	if oldStatesMap[groupID].Action.hasApprovals() {
		personalInfoViewApprovedAt = oldStatesMap[groupID].PersonalInfoViewApprovedAt
		lockMembershipApprovedAt = oldStatesMap[groupID].LockMembershipApprovedAt
		watchApprovedAt = oldStatesMap[groupID].WatchApprovedAt
	} else {
		if approvals[groupID].PersonalInfoViewApproval {
			personalInfoViewApprovedAt = Now()
		}
		if approvals[groupID].LockMembershipApproval {
			lockMembershipApprovedAt = Now()
		}
		if approvals[groupID].WatchApproval {
			watchApprovedAt = Now()
		}
	}
	return personalInfoViewApprovedAt, lockMembershipApprovedAt, watchApprovedAt
}

func insertGroupPendingRequests(dataStore *DataStore, idsToInsertPending map[int64]GroupMembershipAction,
	parentGroupID int64, approvals map[int64]GroupApprovals,
) {
	if len(idsToInsertPending) > 0 {
		insertQuery := `
			INSERT INTO group_pending_requests
				(group_id, member_id, ` + "`type`" + `, personal_info_view_approved,
				 lock_membership_approved, watch_approved, at)`
		valuesTemplate := "(?, ?, ?, ?, ?, ?, NOW(3))"
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsToInsertPending)-1) +
			valuesTemplate // #nosec
		values := make([]interface{}, 0, len(idsToInsertPending)*6) //nolint:mnd // 6 values per row
		for id, groupMembershipAction := range idsToInsertPending {
			values = append(values, parentGroupID, id, groupMembershipAction.PendingType(),
				approvals[id].PersonalInfoViewApproval, approvals[id].LockMembershipApproval,
				approvals[id].WatchApproval)
		}
		mustNotBeError(dataStore.db.Exec(insertQuery, values...).Error)
	}
}

func insertGroupMembershipChanges(dataStore *DataStore, idsChanged map[int64]GroupMembershipAction,
	parentGroupID, performedByUserID int64,
) {
	if len(idsChanged) > 0 {
		insertQuery := "INSERT INTO group_membership_changes (group_id, member_id, action, at, initiator_id)"
		valuesTemplate := "(?, ?, ?, NOW(3), ?)"
		const paramsCount = 4
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsChanged)-1) +
			valuesTemplate // #nosec
		mustNotBeError(dataStore.retryOnDuplicatePrimaryKeyError("group_membership_changes", func(db *DB) error {
			values := make([]interface{}, 0, len(idsChanged)*paramsCount)
			for id, toAction := range idsChanged {
				values = append(values, parentGroupID, id, toAction[strings.LastIndex(string(toAction), ",")+1:], performedByUserID)
			}
			return db.Exec(insertQuery, values...).Error()
		}))
	}
}

func performCyclesChecking(s *DataStore, parentGroupID int64, results GroupGroupTransitionResults, plan *transitionsPlan) {
	if !plan.idsToCheckCycle.IsEmpty() {
		idsToCheckCycleSlice := plan.idsToCheckCycle.Values()
		var cycleIDs []int64
		mustNotBeError(s.GroupAncestors().
			WithSharedWriteLock().
			Where("child_group_id = ? AND ancestor_group_id IN (?)", parentGroupID, idsToCheckCycleSlice).
			Pluck("ancestor_group_id", &cycleIDs).Error())

		deleteIDsFromTransitionPlan(cycleIDs, Cycle, results, plan)
	}
}

func deleteIDsFromTransitionPlan(
	ids []int64, status GroupGroupTransitionResult, results GroupGroupTransitionResults, plan *transitionsPlan,
) {
	for _, groupID := range ids {
		results[groupID] = status
		plan.idsToInsertRelation.Remove(groupID)
		delete(plan.idsToInsertPending, groupID)
		plan.idsToDeletePending.Remove(groupID)
		plan.idsToDeleteRelation.Remove(groupID)
		delete(plan.idsChanged, groupID)
	}
}

type transitionsPlan struct {
	idsToInsertPending  map[int64]GroupMembershipAction
	idsToInsertRelation *golang.Set[int64]
	idsToCheckCycle     *golang.Set[int64]
	idsToDeletePending  *golang.Set[int64]
	idsToDeleteRelation *golang.Set[int64]
	idsChanged          map[int64]GroupMembershipAction
}

func newTransitionsPlan() *transitionsPlan {
	return &transitionsPlan{
		idsToInsertPending:  make(map[int64]GroupMembershipAction),
		idsToInsertRelation: golang.NewSet[int64](),
		idsToCheckCycle:     golang.NewSet[int64](),
		idsToDeletePending:  golang.NewSet[int64](),
		idsToDeleteRelation: golang.NewSet[int64](),
		idsChanged:          make(map[int64]GroupMembershipAction),
	}
}

func buildTransitionsPlan(parentGroupID int64, childGroupIDs []int64, results GroupGroupTransitionResults,
	oldStatesMap map[int64]stateInfo, groupRequiredApprovals *requiredApprovalsAndLimits,
	approvals, approvalsToRequest map[int64]GroupApprovals, action GroupGroupTransitionAction,
) *transitionsPlan {
	plan := newTransitionsPlan()
	for _, childGroupID := range childGroupIDs {
		results[childGroupID] = Invalid
		if childGroupID == parentGroupID {
			continue
		}

		oldState := oldStatesMap[childGroupID]

		toAction, toActionOK := groupGroupTransitionRuleForAction(action).Transitions[oldState.Action]
		if !toActionOK || !checkIfApprovalsAreOK(
			oldState, &toAction, results, groupRequiredApprovals, approvals, approvalsToRequest, action, childGroupID) {
			continue
		}

		buildOneTransition(childGroupID, oldState.Action, toAction, results, plan)
	}
	return plan
}

func checkIfApprovalsAreOK(oldState stateInfo, toAction *GroupMembershipAction,
	results GroupGroupTransitionResults,
	groupRequiredApprovals *requiredApprovalsAndLimits, approvals, approvalsToRequest map[int64]GroupApprovals,
	action GroupGroupTransitionAction, childGroupID int64,
) bool {
	if (!toAction.isActive() || oldState.Action.isActive()) && !toAction.hasApprovals() {
		return true
	}
	ok, approvalsNeeded := approvalsOK(oldState, groupRequiredApprovals, approvals[childGroupID])
	if ok {
		return true
	}
	transitionRule := groupGroupTransitionRuleForAction(action)
	if transitionRule.IfNotEnoughApprovalsDowngradeTo != NoRelation {
		*toAction = transitionRule.IfNotEnoughApprovalsDowngradeTo
		return true
	}

	results[childGroupID] = ApprovalsMissing
	if approvalsNeeded != (GroupApprovals{}) {
		approvalsToRequest[childGroupID] = approvalsNeeded
	}
	return false
}

func buildOneTransition(groupID int64, oldAction, toAction GroupMembershipAction,
	results GroupGroupTransitionResults, plan *transitionsPlan,
) {
	if toAction == oldAction {
		results[groupID] = Unchanged
		return
	}
	if toAction != NoRelation {
		plan.idsChanged[groupID] = toAction
	}
	results[groupID] = Success
	if oldAction.isActive() {
		if !toAction.isActive() {
			plan.idsToDeleteRelation.Add(groupID)
		}
	} else {
		if toAction.isActive() {
			plan.idsToInsertRelation.Add(groupID)
		}
		if toAction.isActive() || toAction.isPending() {
			plan.idsToCheckCycle.Add(groupID)
		}
	}
	if oldAction.isPending() {
		plan.idsToDeletePending.Add(groupID)
	}
	if toAction.isPending() {
		plan.idsToInsertPending[groupID] = toAction
	}
}

func approvalsOK(oldState stateInfo, groupRequiredApprovals *requiredApprovalsAndLimits, approvals GroupApprovals) (
	ok bool, approvalsToRequest GroupApprovals,
) {
	var approvalsToCheck GroupApprovals
	if oldState.Action.hasApprovals() {
		approvalsToCheck.PersonalInfoViewApproval = oldState.PersonalInfoViewApprovedAt != nil
		approvalsToCheck.LockMembershipApproval = oldState.LockMembershipApprovedAt != nil
		approvalsToCheck.WatchApproval = oldState.WatchApprovedAt != nil
	} else {
		approvalsToCheck = approvals
	}
	approvalsToRequest.PersonalInfoViewApproval = groupRequiredApprovals.RequirePersonalInfoAccessApproval &&
		!approvalsToCheck.PersonalInfoViewApproval
	approvalsToRequest.LockMembershipApproval = groupRequiredApprovals.RequireLockMembershipApproval &&
		!approvalsToCheck.LockMembershipApproval
	approvalsToRequest.WatchApproval = groupRequiredApprovals.RequireWatchApproval && !approvalsToCheck.WatchApproval
	return approvalsToRequest == GroupApprovals{}, approvalsToRequest
}
