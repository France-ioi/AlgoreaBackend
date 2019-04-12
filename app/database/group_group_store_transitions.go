package database

import (
	"strings"
)

type GroupGroupType string

const (
	InvitationSent     GroupGroupType = "invitationSent"
	RequestSent        GroupGroupType = "requestSent"
	InvitationAccepted GroupGroupType = "invitationAccepted"
	RequestAccepted    GroupGroupType = "requestAccepted"
	InvitationRefused  GroupGroupType = "invitationRefused"
	RequestRefused     GroupGroupType = "requestRefused"
	Removed            GroupGroupType = "removed"
	Left               GroupGroupType = "left"
	Direct             GroupGroupType = "direct"
	NoRelation         GroupGroupType = ""
)

func (groupType GroupGroupType) IsActive() bool {
	switch groupType {
	case InvitationAccepted, RequestAccepted, Direct:
		return true
	}
	return false
}

type GroupGroupTransitionAction int

const (
	AdminCreatesInvitation GroupGroupTransitionAction = iota
	UserCreatesRequest
	UserAcceptsInvitation
	AdminAcceptsRequest
	UserRefusesInvitation
	AdminRefusesRequest
	// This action marks relations as "removed". It doesn't check if a child is a user or not.
	AdminRemovesUser
	AdminCancelsInvitation
	UserLeavesGroup
	UserCancelsRequest
	// This action creates a new direct relation. It doesn't check if a child is a user or not.
	AdminAddsDirectRelation
	AdminRemovesDirectRelation
)

type groupGroupTransitionRule struct {
	// If a "from" type is listed in "Transitions", but not listed in "UpdateFromType",
	// then a relation is overwritten instead of updating.
	// This field has "GroupGroupType->true" format.
	UpdateFromType map[GroupGroupType]bool
	// Transitions defines all possible transitions for the action. The format is "FromType->ToType".
	// Relations that have "from" type not listed here are considered as invalid for the action.
	Transitions map[GroupGroupType]GroupGroupType
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
		},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationAccepted: Removed,
			RequestAccepted:    Removed,
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
			Direct:             true,
		},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationAccepted: Left,
			RequestAccepted:    Left,
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

type GroupGroupTransitionResults struct {
	Success   map[int64]bool
	Unchanged map[int64]bool
	Invalid   map[int64]bool
	Cycle     map[int64]bool
}

func (s *GroupGroupStore) Transition(action GroupGroupTransitionAction, parentGroupID int64, childGroupIDs []int64) (result *GroupGroupTransitionResults, err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	results := GroupGroupTransitionResults{
		Success:   make(map[int64]bool, len(childGroupIDs)),
		Unchanged: make(map[int64]bool),
		Invalid:   make(map[int64]bool),
		Cycle:     make(map[int64]bool),
	}

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

		idsToUpdate := make(map[int64]GroupGroupType, len(childGroupIDs))
		idsToCheckCycle := make(map[int64]bool, len(childGroupIDs))
		idsToDelete := make(map[int64]bool, len(childGroupIDs))
		idsToInsert := make(map[int64]GroupGroupType, len(childGroupIDs))
		for _, id := range childGroupIDs {
			if id == parentGroupID {
				results.Invalid[id] = true
				continue
			}
			oldType := oldTypesMap[id]
			if toType, toTypeOK := groupGroupTransitionRules[action].Transitions[GroupGroupType(oldType)]; toTypeOK {
				if toType != oldType {
					if !groupGroupTransitionRules[action].UpdateFromType[oldType] {
						idsToDelete[id] = true
					}
					if toType == NoRelation {
						idsToDelete[id] = true
					} else if idsToDelete[id] {
						idsToInsert[id] = toType
					} else {
						idsToUpdate[id] = toType
					}
					results.Success[id] = true
					if toType.IsActive() || oldType == NoRelation {
						idsToCheckCycle[id] = true
					}
				} else {
					results.Unchanged[id] = true
				}
			} else {
				results.Invalid[id] = true
			}
		}

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
				delete(results.Success, idGroup)
				delete(idsToUpdate, idGroup)
				delete(idsToInsert, idGroup)
				results.Cycle[idGroup] = true
			}
		}

		shouldCreateNewAncestors := false
		if len(idsToDelete) > 0 {
			idsToDeleteSlice := make([]int64, 0, len(idsToDelete))
			for id := range idsToDelete {
				idsToDeleteSlice = append(idsToDeleteSlice, id)
			}
			mustNotBeError(s.Delete("idGroupParent = ? AND idGroupChild IN (?)", parentGroupID, idsToDeleteSlice).Error())
			shouldCreateNewAncestors = true
		}

		if len(idsToUpdate) > 0 {
			updateData := map[GroupGroupType][]int64{}
			for id, toType := range idsToUpdate {
				updateData[toType] = append(updateData[toType], id)
				shouldCreateNewAncestors = true
			}
			const updateQuery = `
				UPDATE groups_groups
				SET sType = ?, sStatusDate = NOW()
				WHERE idGroupParent = ? AND idGroupChild IN (?)`
			for toType, ids := range updateData {
				mustNotBeError(s.db.Exec(updateQuery, toType, parentGroupID, ids).Error)
			}
		}

		if len(idsToInsert) > 0 {
			var maxChildOrder struct{ MaxChildOrder int64 }
			mustNotBeError(s.WithWriteLock().
				Select("IFNULL(MAX(iChildOrder), 0)").
				Where("idGroupParent = ?", parentGroupID).Scan(&maxChildOrder).Error())

			insertQuery := "INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, sType, iChildOrder, sStatusDate) VALUES " +
				strings.Repeat("(?, ?, ?, ?, ?, NOW()), ", len(idsToInsert)-1) +
				"(?, ?, ?, ?, ?, NOW())"
			mustNotBeError(s.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
				values := make([]interface{}, 0, len(idsToInsert)*5)
				for id, toType := range idsToInsert {
					maxChildOrder.MaxChildOrder++
					values = append(values, s.NewID(), parentGroupID, id, toType, maxChildOrder.MaxChildOrder)
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
	return &results, nil
}
