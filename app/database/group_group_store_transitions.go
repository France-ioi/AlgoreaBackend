package database

import (
	"strings"
)

// GroupGroupType represents a type of relation between two groups
type GroupGroupType string

const (
	// InvitationSent means there is a pending group admin's invitation for user to join a group
	InvitationSent GroupGroupType = "invitationSent"
	// RequestSent means there is a pending user's request to join a group
	RequestSent GroupGroupType = "requestSent"
	// InvitationAccepted means a user is a member of a group since he has accepted an invitation
	InvitationAccepted GroupGroupType = "invitationAccepted"
	// RequestAccepted means a user is a member of a group since a group admin has accepted his request
	RequestAccepted GroupGroupType = "requestAccepted"
	// InvitationRefused means a user refused an invitation to join a group
	InvitationRefused GroupGroupType = "invitationRefused"
	// JoinedByCode means a user joined a group by the group's code
	JoinedByCode GroupGroupType = "joinedByCode"
	// RequestRefused means an admin refused a user's request to join a group
	RequestRefused GroupGroupType = "requestRefused"
	// Removed means a user was removed from a group
	Removed GroupGroupType = "removed"
	// Left means a user left a group
	Left GroupGroupType = "left"
	// Direct means a direct relation between groups
	Direct GroupGroupType = "direct"
	// NoRelation means there is no row for the group pair in the groups_groups table
	NoRelation GroupGroupType = ""
)

func (groupType GroupGroupType) isActive() bool {
	switch groupType {
	case InvitationAccepted, RequestAccepted, JoinedByCode, Direct:
		return true
	}
	return false
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
	// If a "from" type is listed in "Transitions", but not listed in "UpdateFromType",
	// then a relation is overwritten instead of updating.
	// This field has "GroupGroupType->true" format.
	UpdateFromType map[GroupGroupType]bool
	// Transitions defines all possible transitions for the action. The format is "FromType->ToType".
	// Relations that have "from" type not listed here are considered as invalid for the action.
	Transitions       map[GroupGroupType]GroupGroupType
	SetIDUserInviting bool
}

var groupGroupTransitionRules = map[GroupGroupTransitionAction]groupGroupTransitionRule{
	AdminCreatesInvitation: {
		UpdateFromType: map[GroupGroupType]bool{RequestSent: true},
		Transitions: map[GroupGroupType]GroupGroupType{
			NoRelation:        InvitationSent,
			InvitationSent:    InvitationSent,
			RequestSent:       RequestAccepted,
			InvitationRefused: InvitationSent,
			RequestRefused:    InvitationSent,
			Removed:           InvitationSent,
			Left:              InvitationSent,
		},
		SetIDUserInviting: true,
	},
	UserCreatesRequest: {
		Transitions: map[GroupGroupType]GroupGroupType{
			NoRelation:        RequestSent,
			RequestSent:       RequestSent,
			InvitationRefused: RequestSent,
			RequestRefused:    RequestSent,
			Removed:           RequestSent,
			Left:              RequestSent,
		},
	},
	UserCreatesAcceptedRequest: {
		Transitions: map[GroupGroupType]GroupGroupType{
			NoRelation:        RequestAccepted,
			RequestSent:       RequestAccepted,
			InvitationSent:    RequestAccepted,
			InvitationRefused: RequestAccepted,
			RequestRefused:    RequestAccepted,
			RequestAccepted:   RequestAccepted,
			Removed:           RequestAccepted,
			Left:              RequestAccepted,
		},
	},
	UserJoinsGroupByCode: {
		Transitions: map[GroupGroupType]GroupGroupType{
			NoRelation:        JoinedByCode,
			RequestSent:       JoinedByCode,
			InvitationRefused: JoinedByCode,
			InvitationSent:    JoinedByCode,
			RequestRefused:    JoinedByCode,
			Removed:           JoinedByCode,
			Left:              JoinedByCode,
		},
	},
	UserAcceptsInvitation: {
		UpdateFromType: map[GroupGroupType]bool{InvitationSent: true},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationSent:     InvitationAccepted,
			InvitationAccepted: InvitationAccepted,
		},
	},
	AdminAcceptsRequest: {
		UpdateFromType: map[GroupGroupType]bool{RequestSent: true},
		Transitions: map[GroupGroupType]GroupGroupType{
			RequestSent:     RequestAccepted,
			RequestAccepted: RequestAccepted,
		},
	},
	UserRefusesInvitation: {
		UpdateFromType: map[GroupGroupType]bool{InvitationSent: true},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationSent:    InvitationRefused,
			InvitationRefused: InvitationRefused,
		},
	},
	AdminRefusesRequest: {
		UpdateFromType: map[GroupGroupType]bool{RequestSent: true},
		Transitions: map[GroupGroupType]GroupGroupType{
			RequestSent:    RequestRefused,
			RequestRefused: RequestRefused,
		},
	},
	AdminRemovesUser: {
		UpdateFromType: map[GroupGroupType]bool{
			InvitationAccepted: true,
			RequestAccepted:    true,
			JoinedByCode:       true,
		},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationAccepted: Removed,
			RequestAccepted:    Removed,
			JoinedByCode:       Removed,
			Removed:            Removed,
		},
	},
	AdminCancelsInvitation: {
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationSent: NoRelation,
		},
	},
	UserLeavesGroup: {
		UpdateFromType: map[GroupGroupType]bool{
			InvitationAccepted: true,
			RequestAccepted:    true,
			JoinedByCode:       true,
			Direct:             true,
		},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationAccepted: Left,
			RequestAccepted:    Left,
			JoinedByCode:       Left,
			Left:               Left,
			Direct:             Left,
		},
	},
	UserCancelsRequest: {
		Transitions: map[GroupGroupType]GroupGroupType{
			RequestSent: NoRelation,
		},
	},
	// This one is here for consistency purpose only.
	// GroupGroupStore.CreateRelation() is more effective when we need to create just one relation.
	AdminAddsDirectRelation: {
		Transitions: map[GroupGroupType]GroupGroupType{
			NoRelation:         Direct,
			InvitationSent:     Direct,
			RequestSent:        Direct,
			InvitationAccepted: Direct,
			RequestAccepted:    Direct,
			JoinedByCode:       Direct,
			InvitationRefused:  Direct,
			RequestRefused:     Direct,
			Removed:            Direct,
			Left:               Direct,
			Direct:             Direct,
		},
	},
	AdminRemovesDirectRelation: {
		Transitions: map[GroupGroupType]GroupGroupType{
			Direct:     NoRelation,
			NoRelation: NoRelation,
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
		type idWithType struct {
			ChildGroupID int64          `gorm:"column:idGroupChild"`
			Type         GroupGroupType `gorm:"column:sType"`
		}
		var oldTypes []idWithType
		mustNotBeError(s.WithWriteLock().
			Select("idGroupChild, sType").
			Where("idGroupParent = ? AND idGroupChild IN (?)", parentGroupID, childGroupIDs).
			Scan(&oldTypes).Error())

		oldTypesMap := make(map[int64]GroupGroupType, len(childGroupIDs))
		for _, oldType := range oldTypes {
			oldTypesMap[oldType.ChildGroupID] = oldType.Type
		}

		idsToInsert, idsToUpdate, idsToCheckCycle, idsToDelete := buildTransitionsPlan(
			parentGroupID, childGroupIDs, results, oldTypesMap, action)

		performCyclesChecking(s.DataStore, idsToCheckCycle, parentGroupID, results, idsToInsert, idsToUpdate, idsToDelete)

		setIDUserInviting := groupGroupTransitionRules[action].SetIDUserInviting
		shouldCreateNewAncestors := false
		if len(idsToDelete) > 0 {
			idsToDeleteSlice := make([]int64, 0, len(idsToDelete))
			for id := range idsToDelete {
				idsToDeleteSlice = append(idsToDeleteSlice, id)
			}
			mustNotBeError(s.Delete("idGroupParent = ? AND idGroupChild IN (?)", parentGroupID, idsToDeleteSlice).Error())
			shouldCreateNewAncestors = true
		}

		shouldCreateNewAncestors = performTransitionUpdate(s, idsToUpdate, parentGroupID, setIDUserInviting, performedByUserID) ||
			shouldCreateNewAncestors

		if len(idsToInsert) > 0 {
			var maxChildOrder struct{ MaxChildOrder int64 }
			mustNotBeError(s.WithWriteLock().
				Select("IFNULL(MAX(iChildOrder), 0)").
				Where("idGroupParent = ?", parentGroupID).Scan(&maxChildOrder).Error())

			insertQuery := "INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, sType, iChildOrder, sStatusDate"
			valuesTemplate := "(?, ?, ?, ?, ?, NOW()"
			paramsCount := 5
			if setIDUserInviting {
				insertQuery += ", idUserInviting"
				valuesTemplate += ", ?"
				paramsCount++
			}
			valuesTemplate += ")"
			insertQuery += ") VALUES " +
				strings.Repeat(valuesTemplate+", ", len(idsToInsert)-1) +
				valuesTemplate // #nosec
			mustNotBeError(s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
				values := make([]interface{}, 0, len(idsToInsert)*paramsCount)
				for id, toType := range idsToInsert {
					maxChildOrder.MaxChildOrder++
					values = append(values, s.NewID(), parentGroupID, id, toType, maxChildOrder.MaxChildOrder)
					if setIDUserInviting {
						values = append(values, performedByUserID)
					}
					shouldCreateNewAncestors = true
				}
				return s.db.Exec(insertQuery, values...).Error
			}))
		}

		if shouldCreateNewAncestors {
			s.createNewAncestors()
		}
		return nil
	}))
	return results, nil
}

func performCyclesChecking(s *DataStore, idsToCheckCycle map[int64]bool, parentGroupID int64,
	results GroupGroupTransitionResults, idsToInsert, idsToUpdate map[int64]GroupGroupType,
	idsToDelete map[int64]bool) {
	if len(idsToCheckCycle) > 0 {
		idsToCheckCycleSlice := make([]int64, 0, len(idsToCheckCycle))
		for id := range idsToCheckCycle {
			idsToCheckCycleSlice = append(idsToCheckCycleSlice, id)
		}
		var cycleIDs []map[string]interface{}
		mustNotBeError(s.GroupAncestors().
			WithWriteLock().
			Select("idGroupAncestor AS idGroup").
			Where("idGroupChild = ? AND idGroupAncestor IN (?)", parentGroupID, idsToCheckCycleSlice).
			ScanIntoSliceOfMaps(&cycleIDs).Error())

		for _, cycleID := range cycleIDs {
			idGroup := cycleID["idGroup"].(int64)
			results[idGroup] = Cycle
			delete(idsToUpdate, idGroup)
			delete(idsToInsert, idGroup)
			delete(idsToDelete, idGroup)
		}
	}
}

func performTransitionUpdate(
	s *GroupGroupStore, idsToUpdate map[int64]GroupGroupType, parentGroupID int64,
	setIDUserInviting bool, performedByUserID int64) (shouldCreateNewAncestors bool) {
	if len(idsToUpdate) > 0 {
		updateData := map[GroupGroupType][]int64{}
		for id, toType := range idsToUpdate {
			updateData[toType] = append(updateData[toType], id)
			shouldCreateNewAncestors = true
		}

		updateQuery := `
			UPDATE groups_groups
			SET sType = ?, sStatusDate = NOW()`
		if setIDUserInviting {
			updateQuery += ", idUserInviting = ?"
		}
		updateQuery += "\nWHERE idGroupParent = ? AND idGroupChild IN (?)"
		for toType, ids := range updateData {
			parameters := make([]interface{}, 0, 5)
			parameters = append(parameters, toType)
			if setIDUserInviting {
				parameters = append(parameters, performedByUserID)
			}
			parameters = append(parameters, parentGroupID, ids)
			mustNotBeError(s.db.Exec(updateQuery, parameters...).Error)
		}
	}
	return shouldCreateNewAncestors
}

func buildTransitionsPlan(parentGroupID int64, childGroupIDs []int64, results GroupGroupTransitionResults,
	oldTypesMap map[int64]GroupGroupType, action GroupGroupTransitionAction,
) (idsToInsert, idsToUpdate map[int64]GroupGroupType, idsToCheckCycle, idsToDelete map[int64]bool) {
	idsToUpdate = make(map[int64]GroupGroupType, len(childGroupIDs))
	idsToCheckCycle = make(map[int64]bool, len(childGroupIDs))
	idsToDelete = make(map[int64]bool, len(childGroupIDs))
	idsToInsert = make(map[int64]GroupGroupType, len(childGroupIDs))
	for _, id := range childGroupIDs {
		results[id] = Invalid
		if id == parentGroupID {
			continue
		}
		oldType := oldTypesMap[id]
		if toType, toTypeOK := groupGroupTransitionRules[action].Transitions[oldType]; toTypeOK {
			buildOneTransition(id, oldType, toType, action, results, idsToInsert, idsToUpdate, idsToCheckCycle, idsToDelete)
		}
	}
	return idsToInsert, idsToUpdate, idsToCheckCycle, idsToDelete
}

func buildOneTransition(id int64, oldType, toType GroupGroupType, action GroupGroupTransitionAction,
	results GroupGroupTransitionResults,
	idsToInsert, idsToUpdate map[int64]GroupGroupType, idsToCheckCycle, idsToDelete map[int64]bool) {
	if toType != oldType {
		results[id] = Success
		if !groupGroupTransitionRules[action].UpdateFromType[oldType] {
			idsToDelete[id] = true
		}
		switch {
		case toType == NoRelation:
			idsToDelete[id] = true
		case idsToDelete[id]:
			idsToInsert[id] = toType
		default:
			idsToUpdate[id] = toType
		}
		if toType.isActive() || oldType == NoRelation {
			idsToCheckCycle[id] = true
		}
	} else {
		results[id] = Unchanged
	}
}
