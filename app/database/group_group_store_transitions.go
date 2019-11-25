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
	// AddedDirectly means a user was added into a group directly
	AddedDirectly GroupMembershipAction = "added_directly"
	// NoRelation means there is no row for the group pair in the groups_groups/group_pending_requests tables
	NoRelation GroupMembershipAction = ""
)

func (groupMembershipAction GroupMembershipAction) isActive() bool {
	switch groupMembershipAction {
	case InvitationAccepted, JoinRequestAccepted, JoinedByCode, AddedDirectly:
		return true
	}
	return false
}

func (groupMembershipAction GroupMembershipAction) isPending() bool {
	switch groupMembershipAction {
	case InvitationCreated, JoinRequestCreated:
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
	}
	panic("groupMembershipAction should be of pending kind in PendingType()")
}

// GroupGroupTransitionAction represents a groups_groups relation transition action
type GroupGroupTransitionAction int

const (
	// AdminCreatesInvitation means a group admin invites new users to the group
	AdminCreatesInvitation GroupGroupTransitionAction = iota
	// UserCreatesRequest means a user sends request to become a group member
	UserCreatesRequest
	// UserCreatesAcceptedRequest means a user adds himself into a group that he owns
	// It doesn't check if the user owns the group (a calling service should check that)
	UserCreatesAcceptedRequest
	// UserAcceptsInvitation means a user accepts a group invitation
	UserAcceptsInvitation
	// AdminAcceptsRequest means a group admin accepts a request
	AdminAcceptsRequest
	// UserRefusesInvitation means a user refuses a group invitation
	UserRefusesInvitation
	// AdminRefusesRequest means a group admin refuses a request to join the group
	AdminRefusesRequest
	// AdminRemovesUser means a group admin removes a user from a group. It marks relations as "removed".
	// It doesn't check if a child is a user or not.
	AdminRemovesUser
	// AdminCancelsInvitation means a group admin cancels an invitation
	AdminCancelsInvitation
	// UserLeavesGroup means a user leaves a group
	UserLeavesGroup
	// UserCancelsRequest means a user cancels his request to join a group
	UserCancelsRequest
	// AdminAddsDirectRelation means a group admin creates a direct relation between groups.
	// It creates a new direct relation. It doesn't check if a child is a user or not.
	AdminAddsDirectRelation
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
			NoRelation:         InvitationCreated,
			InvitationCreated:  InvitationCreated,
			JoinRequestCreated: JoinRequestAccepted,
		},
	},
	UserCreatesRequest: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			NoRelation:         JoinRequestCreated,
			JoinRequestCreated: JoinRequestCreated,
		},
	},
	UserCreatesAcceptedRequest: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			NoRelation:         JoinRequestAccepted,
			JoinRequestCreated: JoinRequestAccepted,
			InvitationCreated:  JoinRequestAccepted,
		},
	},
	UserJoinsGroupByCode: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			NoRelation:         JoinedByCode,
			JoinRequestCreated: JoinedByCode,
			InvitationCreated:  JoinedByCode,
		},
	},
	UserAcceptsInvitation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			InvitationCreated: InvitationAccepted,
		},
	},
	AdminAcceptsRequest: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			JoinRequestCreated: JoinRequestAccepted,
		},
	},
	UserRefusesInvitation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			InvitationCreated: InvitationRefused,
		},
	},
	AdminRefusesRequest: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			JoinRequestCreated: JoinRequestRefused,
		},
	},
	AdminRemovesUser: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			AddedDirectly: Removed,
		},
	},
	AdminCancelsInvitation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			InvitationCreated: InvitationWithdrawn,
		},
	},
	UserLeavesGroup: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			InvitationAccepted:  Left,
			JoinRequestAccepted: Left,
			JoinedByCode:        Left,
			AddedDirectly:       Left,
		},
	},
	UserCancelsRequest: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			JoinRequestCreated: JoinRequestWithdrawn,
		},
	},
	// This one is here for consistency purpose only.
	// GroupGroupStore.CreateRelation() is more effective when we need to create just one relation.
	AdminAddsDirectRelation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			NoRelation:          AddedDirectly,
			InvitationCreated:   AddedDirectly,
			JoinRequestCreated:  AddedDirectly,
			InvitationAccepted:  AddedDirectly,
			JoinRequestAccepted: AddedDirectly,
			JoinedByCode:        AddedDirectly,
			AddedDirectly:       AddedDirectly,
		},
	},
	AdminRemovesDirectRelation: {
		Transitions: map[GroupMembershipAction]GroupMembershipAction{
			AddedDirectly: NoRelation,
			NoRelation:    NoRelation,
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

// Transition performs a groups_groups relation transition according to groupGroupTransitionRules
func (s *GroupGroupStore) Transition(action GroupGroupTransitionAction,
	parentGroupID int64, childGroupIDs []int64, performedByUserID int64) (result GroupGroupTransitionResults, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	results := GroupGroupTransitionResults(make(map[int64]GroupGroupTransitionResult, len(childGroupIDs)))

	mustNotBeError(s.WithNamedLock(s.tableName, groupsRelationsLockTimeout, func(dataStore *DataStore) error {
		type idWithAction struct {
			ChildGroupID int64
			Action       GroupMembershipAction
		}
		var oldActions []idWithAction

		mustNotBeError(dataStore.Raw("(? FOR UPDATE) UNION (? FOR UPDATE)",
			dataStore.GroupGroups().
				Select("child_group_id, 'added_directly' AS `type`").
				Where("parent_group_id = ? AND child_group_id IN (?)", parentGroupID, childGroupIDs).QueryExpr(),
			dataStore.GroupPendingRequests().
				Select(`
					member_id,
					CASE type
						WHEN 'invitation' THEN 'invitation_created'
						WHEN 'join_request' THEN 'join_request_created'
						ELSE type
					END`).
				Where("group_id = ? AND member_id IN (?)", parentGroupID, childGroupIDs).QueryExpr()).
			Scan(&oldActions).Error())

		oldActionsMap := make(map[int64]GroupMembershipAction, len(childGroupIDs))
		for _, oldAction := range oldActions {
			oldActionsMap[oldAction.ChildGroupID] = oldAction.Action
		}

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

		insertGroupPendingRequests(dataStore, idsToInsertPending, parentGroupID)

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

func insertGroupPendingRequests(dataStore *DataStore, idsToInsertPending map[int64]GroupMembershipAction, parentGroupID int64) {
	if len(idsToInsertPending) > 0 {
		insertQuery := "INSERT INTO group_pending_requests (group_id, member_id, `type`)"
		valuesTemplate := "(?, ?, ?)"
		insertQuery += " VALUES " +
			strings.Repeat(valuesTemplate+", ", len(idsToInsertPending)-1) +
			valuesTemplate // #nosec
		values := make([]interface{}, 0, len(idsToInsertPending)*3)
		for id, groupMembershipAction := range idsToInsertPending {
			values = append(values, parentGroupID, id, groupMembershipAction.PendingType())
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
				values = append(values, parentGroupID, id, toAction, performedByUserID)
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
			idsToDeleteRelation[id] = true
		} else {
			idsToDeletePending[id] = true
		}
		switch {
		case toAction.isActive():
			idsToInsertRelation[id] = true
		case toAction.isPending():
			idsToInsertPending[id] = toAction
		}
		if toAction.isActive() || toAction.isPending() {
			idsToCheckCycle[id] = true
		}
	} else {
		results[id] = Unchanged
	}
}
