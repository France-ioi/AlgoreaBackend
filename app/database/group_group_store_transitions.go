package database

import (
	"strings"
)

// GroupMembershipAction represents an action that changes relation between two groups
type GroupMembershipAction string

const (
	// InvitationCreated means a pending group admin's invitation for user to join a group was created
	InvitationCreated GroupMembershipAction = "invitation_created"
	// JoinRequestCreated means a pending user's request to join a group was created
	JoinRequestCreated GroupMembershipAction = "join_request_created"
	// InvitationAccepted means a user became a member of a group by accepting an invitation
	InvitationAccepted GroupMembershipAction = "invitation_accepted"
	// JoinRequestAccepted means a user became a member of a group since a group admin accepted his request
	JoinRequestAccepted GroupMembershipAction = "join_request_accepted"
	// LeaveRequestAccepted means a user left a group since a group admin accepted his leave request
	LeaveRequestAccepted GroupMembershipAction = "leave_request_accepted"
	// InvitationRefused means a user refused an invitation to join a group
	InvitationRefused GroupMembershipAction = "invitation_refused"
	// InvitationWithdrawn means an admin withdrew his invitation to join a group
	InvitationWithdrawn GroupMembershipAction = "invitation_withdrawn"
	// JoinedByCode means a user joined a group by the group's code
	JoinedByCode GroupMembershipAction = "joined_by_code"
	// JoinRequestRefused means an admin refused a user's request to join a group
	JoinRequestRefused GroupMembershipAction = "join_request_refused"
	// JoinRequestWithdrawn means a user withdrew his request to join a group
	JoinRequestWithdrawn GroupMembershipAction = "join_request_withdrawn"
	// Removed means a user was removed from a group
	Removed GroupMembershipAction = "removed"
	// Left means a user left a group
	Left GroupMembershipAction = "left"
	// IsMember means a user is a member of a group
	IsMember GroupMembershipAction = "is_member"
	// LeaveRequestCreated means a pending user's request to leave a group was created
	LeaveRequestCreated GroupMembershipAction = "is_member,leave_request_created"
	// LeaveRequestExpired means a pending user's leave request for an expired membership
	LeaveRequestExpired GroupMembershipAction = "leave_request_created"
	// LeaveRequestRefused means a manager refused a user's request to leave a group
	LeaveRequestRefused GroupMembershipAction = "leave_request_refused"
	// LeaveRequestWithdrawn means a user withdrew his request to leave a group
	LeaveRequestWithdrawn GroupMembershipAction = "leave_request_withdrawn"
	// NoRelation means there is no row for the group pair in the groups_groups/group_pending_requests tables
	NoRelation GroupMembershipAction = ""
)

func (groupMembershipAction GroupMembershipAction) isActive() bool {
	switch groupMembershipAction {
	case InvitationAccepted, JoinRequestAccepted, JoinedByCode, IsMember,
		LeaveRequestCreated, LeaveRequestWithdrawn, LeaveRequestRefused:
		return true
	}
	return false
}

func (groupMembershipAction GroupMembershipAction) isPending() bool {
	switch groupMembershipAction {
	case InvitationCreated, JoinRequestCreated, LeaveRequestCreated, LeaveRequestExpired:
		return true
	}
	return false
}

// PendingType converts the GroupMembershipAction into `group_pending_requests.type`
func (groupMembershipAction GroupMembershipAction) PendingType() string {
	switch groupMembershipAction {
	case InvitationCreated:
		return "invitation"
	case JoinRequestCreated:
		return "join_request"
	case LeaveRequestCreated:
		return "leave_request"
	}
	panic("groupMembershipAction should be of pending kind in PendingType()")
}

// GroupGroupTransitionAction represents a groups_groups relation transition action
type GroupGroupTransitionAction int

const (
	// AdminCreatesInvitation means a group admin invites new users to the group
	AdminCreatesInvitation GroupGroupTransitionAction = iota
	// UserCreatesJoinRequest means a user sends a request to become a group member
	UserCreatesJoinRequest
	// UserCreatesAcceptedJoinRequest means a user adds himself into a group that he owns
	// It doesn't check if the user owns the group / all needed approvals are given (a calling service should check that)
	UserCreatesAcceptedJoinRequest
	// UserAcceptsInvitation means a user accepts a group invitation
	UserAcceptsInvitation
	// AdminAcceptsJoinRequest means a group admin accepts a request to join a group
	AdminAcceptsJoinRequest
	// AdminAcceptsLeaveRequest means a group admin accepts a request to leave a group
	AdminAcceptsLeaveRequest
	// AdminRefusesLeaveRequest means a group admin refuses a request to leave a group
	AdminRefusesLeaveRequest
	// UserRefusesInvitation means a user refuses a group invitation
	UserRefusesInvitation
	// AdminRefusesJoinRequest means a group admin refuses a request to join the group
	AdminRefusesJoinRequest
	// AdminRemovesUser means a group admin removes a user from a group. It marks relations as "removed".
	// It doesn't check if a child is a user or not.
	AdminRemovesUser
	// AdminWithdrawsInvitation means a group admin withdraws an invitation
	AdminWithdrawsInvitation
	// UserLeavesGroup means a user leaves a group
	UserLeavesGroup
	// UserCreatesLeaveRequest means a user sends a request to leave a group
	// We don't check that groups.require_lock_membership_approval_until & groups_groups.lock_membership_approved_at
	// are not null (a calling service should check that by itself)
	UserCreatesLeaveRequest
	// UserCancelsJoinRequest means a user cancels his request to join a group
	UserCancelsJoinRequest
	// UserCancelsLeaveRequest means a user cancels his request to leave a group
	UserCancelsLeaveRequest
	// AdminRemovesDirectRelation removes a direct relation
	AdminRemovesDirectRelation
	// UserJoinsGroupByCode means a user joins a group using a group's code
	// We don't check the code here (a calling service should check the code by itself)
	UserJoinsGroupByCode
)

type groupGroupTransitionRule struct {
	// Transitions defines all possible transitions for the action. The format is "FromAction->ToAction".
	// Relations that have "from" action not listed here are considered as invalid.
	Transitions map[GroupMembershipAction]GroupMembershipAction
}

var groupGroupTransitionRules = map[GroupGroupTransitionAction]groupGroupTransitionRule{
	AdminCreatesInvitation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			NoRelation:          InvitationCreated,
			InvitationCreated:   InvitationCreated,
			JoinRequestCreated:  JoinRequestAccepted,
			LeaveRequestExpired: InvitationCreated,
		},
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
}

// GroupGroupTransitionResult is an enum{cycle, invalid, success, unchanged}
type GroupGroupTransitionResult string

const (
	// Cycle means that the transition wasn't performed because it would create a cycle in groups_groups graph
	Cycle GroupGroupTransitionResult = "cycle"
	// Invalid means that the transition is impossible
	Invalid GroupGroupTransitionResult = "invalid"
	// Success means that the transition was performed successfully
	Success GroupGroupTransitionResult = "success"
	// Unchanged means that the transition has been already performed
	Unchanged GroupGroupTransitionResult = "unchanged"
)

// GroupGroupTransitionResults represents results of mass transition (format: map{ id -> GroupGroupTransitionResult })
type GroupGroupTransitionResults map[int64]GroupGroupTransitionResult

// GroupApprovals represents all the approvals that can be given by a user to the group managers
type GroupApprovals struct {
	PersonalInfoViewApproval bool
	LockMembershipApproval   bool
	WatchApproval            bool
}

// FromString initializes GroupApprovals from the given comma-separated list of approvals
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

// Transition performs a groups_groups relation transition according to groupGroupTransitionRules
func (s *GroupGroupStore) Transition(action GroupGroupTransitionAction,
	parentGroupID int64, childGroupIDs []int64, approvals map[int64]GroupApprovals,
	performedByUserID int64) (result GroupGroupTransitionResults, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	results := GroupGroupTransitionResults(make(map[int64]GroupGroupTransitionResult, len(childGroupIDs)))

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(dataStore *DataStore) error {
		type idWithAction struct {
			ChildGroupID int64
			Action       GroupMembershipAction
		}
		var oldActions []idWithAction

		// Here we get current states for each childGroupID:
		// the current state can be one of
		// ("", "invitation_created", "join_request_created", "is_member", "is_member,leave_request_created")
		// where "is_member" means that childGroupID is a member of the parentGroupID
		mustNotBeError(
			dataStore.Raw(`
				SELECT child_group_id, GROUP_CONCAT(action) AS action
					FROM ((? FOR UPDATE) UNION (? FOR UPDATE)) AS statuses
					GROUP BY child_group_id`,
				dataStore.ActiveGroupGroups().
					Select("child_group_id, 'is_member' AS `action`").
					Where("parent_group_id = ? AND child_group_id IN (?)", parentGroupID, childGroupIDs).QueryExpr(),
				dataStore.GroupPendingRequests().
					Select(`
					member_id,
					CASE type
						WHEN 'invitation' THEN 'invitation_created'
						WHEN 'join_request' THEN 'join_request_created'
						WHEN 'leave_request' THEN 'leave_request_created'
						ELSE type
					END`).
					Where("group_id = ? AND member_id IN (?)", parentGroupID, childGroupIDs).QueryExpr()).
				Scan(&oldActions).Error())

		oldActionsMap := make(map[int64]GroupMembershipAction, len(childGroupIDs))
		for _, oldAction := range oldActions {
			oldActionsMap[oldAction.ChildGroupID] = oldAction.Action
		}

		// build the transition plan depending on the current states (oldActionsMap)
		idsToInsertPending, idsToInsertRelation, idsToCheckCycle, idsToDeletePending,
			idsToDeleteRelation, idsChanged := buildTransitionsPlan(
			parentGroupID, childGroupIDs, results, oldActionsMap, action)

		performCyclesChecking(dataStore, idsToCheckCycle, parentGroupID, results, idsToInsertPending, idsToInsertRelation,
			idsToDeletePending, idsToDeleteRelation, idsChanged)

		shouldCreateNewAncestors := false
		if len(idsToDeletePending) > 0 {
			idsToDeleteSlice := make([]int64, 0, len(idsToDeletePending))
			for id := range idsToDeletePending {
				idsToDeleteSlice = append(idsToDeleteSlice, id)
			}
			mustNotBeError(dataStore.GroupPendingRequests().Delete("group_id = ? AND member_id IN (?)", parentGroupID, idsToDeleteSlice).Error())
		}
		if len(idsToDeleteRelation) > 0 {
			idsToDeleteSlice := make([]int64, 0, len(idsToDeleteRelation))
			for id := range idsToDeleteRelation {
				idsToDeleteSlice = append(idsToDeleteSlice, id)
			}
			mustNotBeError(dataStore.GroupGroups().Delete("parent_group_id = ? AND child_group_id IN (?)", parentGroupID, idsToDeleteSlice).Error())
			shouldCreateNewAncestors = true
		}

		insertGroupPendingRequests(dataStore, idsToInsertPending, parentGroupID, approvals)

		if len(idsToInsertRelation) > 0 {
			var maxChildOrder struct{ MaxChildOrder int64 }
			mustNotBeError(dataStore.WithWriteLock().
				Select("IFNULL(MAX(child_order), 0)").
				Where("parent_group_id = ?", parentGroupID).Scan(&maxChildOrder).Error())

			insertQuery := "INSERT INTO groups_groups (id, parent_group_id, child_group_id, child_order)"
			valuesTemplate := "(?, ?, ?, ?)"
			insertQuery += " VALUES " +
				strings.Repeat(valuesTemplate+", ", len(idsToInsertRelation)-1) +
				valuesTemplate // #nosec
			insertQuery += " ON DUPLICATE KEY UPDATE expires_at = '9999-12-31 23:59:59'"
			mustNotBeError(dataStore.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
				values := make([]interface{}, 0, len(idsToInsertRelation)*4)
				for id := range idsToInsertRelation {
					maxChildOrder.MaxChildOrder++
					values = append(values, NewDataStore(db).NewID(), parentGroupID, id, maxChildOrder.MaxChildOrder)
					shouldCreateNewAncestors = true
				}
				return db.Exec(insertQuery, values...).Error()
			}))
		}

		insertGroupMembershipChanges(dataStore, idsChanged, parentGroupID, performedByUserID)

		if shouldCreateNewAncestors {
			dataStore.GroupGroups().createNewAncestors()
		}
		return nil
	}))
	return results, nil
}

func insertGroupPendingRequests(dataStore *DataStore, idsToInsertPending map[int64]GroupMembershipAction,
	parentGroupID int64, approvals map[int64]GroupApprovals) {
	if len(idsToInsertPending) > 0 {
		insertQuery := `
			INSERT INTO group_pending_requests
				(group_id, member_id, ` + "`type`" + `, personal_info_view_approved,
				 lock_membership_approved, watch_approved)`
		valuesTemplate := "(?, ?, ?, ?, ?, ?)"
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsToInsertPending)-1) +
			valuesTemplate // #nosec
		values := make([]interface{}, 0, len(idsToInsertPending)*6)
		for id, groupMembershipAction := range idsToInsertPending {
			values = append(values, parentGroupID, id, groupMembershipAction.PendingType(),
				approvals[id].PersonalInfoViewApproval, approvals[id].LockMembershipApproval,
				approvals[id].WatchApproval)
		}
		mustNotBeError(dataStore.db.Exec(insertQuery, values...).Error)
	}
}

func insertGroupMembershipChanges(dataStore *DataStore, idsChanged map[int64]GroupMembershipAction,
	parentGroupID, performedByUserID int64) {
	if len(idsChanged) > 0 {
		insertQuery := "INSERT INTO group_membership_changes (group_id, member_id, action, at, initiator_id)"
		valuesTemplate := "(?, ?, ?, NOW(3), ?)"
		const paramsCount = 4
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsChanged)-1) +
			valuesTemplate // #nosec
		mustNotBeError(dataStore.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
			values := make([]interface{}, 0, len(idsChanged)*paramsCount)
			for id, toAction := range idsChanged {
				values = append(values, parentGroupID, id, toAction[strings.LastIndex(string(toAction), ",")+1:], performedByUserID)
			}
			return dataStore.db.Exec(insertQuery, values...).Error
		}))
	}
}

func performCyclesChecking(s *DataStore, idsToCheckCycle map[int64]bool, parentGroupID int64,
	results GroupGroupTransitionResults, idsToInsertPending map[int64]GroupMembershipAction, idsToInsertRelation,
	idsToDeletePending, idsToDeleteRelation map[int64]bool, idsChanged map[int64]GroupMembershipAction) {
	if len(idsToCheckCycle) > 0 {
		idsToCheckCycleSlice := make([]int64, 0, len(idsToCheckCycle))
		for id := range idsToCheckCycle {
			idsToCheckCycleSlice = append(idsToCheckCycleSlice, id)
		}
		var cycleIDs []map[string]interface{}
		mustNotBeError(s.GroupAncestors().
			WithWriteLock().
			Select("ancestor_group_id AS group_id").
			Where("child_group_id = ? AND ancestor_group_id IN (?)", parentGroupID, idsToCheckCycleSlice).
			ScanIntoSliceOfMaps(&cycleIDs).Error())

		for _, cycleID := range cycleIDs {
			groupID := cycleID["group_id"].(int64)
			results[groupID] = Cycle
			delete(idsToInsertRelation, groupID)
			delete(idsToInsertPending, groupID)
			delete(idsToDeletePending, groupID)
			delete(idsToDeleteRelation, groupID)
			delete(idsChanged, groupID)
		}
	}
}

func buildTransitionsPlan(parentGroupID int64, childGroupIDs []int64, results GroupGroupTransitionResults,
	oldActionsMap map[int64]GroupMembershipAction, action GroupGroupTransitionAction,
) (idsToInsertPending map[int64]GroupMembershipAction, idsToInsertRelation, idsToCheckCycle,
	idsToDeletePending, idsToDeleteRelation map[int64]bool, idsChanged map[int64]GroupMembershipAction) {
	idsToCheckCycle = make(map[int64]bool, len(childGroupIDs))
	idsToDeletePending = make(map[int64]bool, len(childGroupIDs))
	idsToDeleteRelation = make(map[int64]bool, len(childGroupIDs))
	idsToInsertPending = make(map[int64]GroupMembershipAction, len(childGroupIDs))
	idsToInsertRelation = make(map[int64]bool, len(childGroupIDs))
	idsChanged = make(map[int64]GroupMembershipAction, len(childGroupIDs))
	for _, id := range childGroupIDs {
		results[id] = Invalid
		if id == parentGroupID {
			continue
		}
		oldAction := oldActionsMap[id]
		if toAction, toActionOK := groupGroupTransitionRules[action].Transitions[oldAction]; toActionOK {
			buildOneTransition(id, oldAction, toAction, results, idsToInsertPending, idsToInsertRelation, idsToCheckCycle,
				idsToDeletePending, idsToDeleteRelation, idsChanged)
		}
	}
	return idsToInsertPending, idsToInsertRelation, idsToCheckCycle, idsToDeletePending, idsToDeleteRelation, idsChanged
}

func buildOneTransition(id int64, oldAction, toAction GroupMembershipAction,
	results GroupGroupTransitionResults,
	idsToInsertPending map[int64]GroupMembershipAction, idsToInsertRelation, idsToCheckCycle, idsToDeletePending,
	idsToDeleteRelation map[int64]bool, idsChanged map[int64]GroupMembershipAction) {
	if toAction != oldAction {
		if toAction != NoRelation {
			idsChanged[id] = toAction
		}
		results[id] = Success
		if oldAction.isActive() {
			if !toAction.isActive() {
				idsToDeleteRelation[id] = true
			}
		} else {
			if toAction.isActive() {
				idsToInsertRelation[id] = true
			}
			if toAction.isActive() || toAction.isPending() {
				idsToCheckCycle[id] = true
			}
		}
		if oldAction.isPending() {
			idsToDeletePending[id] = true
		}
		if toAction.isPending() {
			idsToInsertPending[id] = toAction
		}
	} else {
		results[id] = Unchanged
	}
}
