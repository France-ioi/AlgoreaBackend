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
	AdminRemovesChild
	AdminCancelsInvitation
	UserLeavesGroup
	UserCancelsRequest
	AdminAddsDirectRelation
)

type groupGroupTransitionRule struct {
	UpdateFromType map[GroupGroupType]bool
	Transitions    map[GroupGroupType]GroupGroupType
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
	AdminRemovesChild: {
		UpdateFromType: map[GroupGroupType]bool{
			InvitationAccepted: true,
			RequestAccepted:    true,
			InvitationRefused:  true,
			RequestRefused:     true,
			Direct:             true,
		},
		Transitions: map[GroupGroupType]GroupGroupType{
			InvitationAccepted: Removed,
			RequestAccepted:    Removed,
			InvitationRefused:  Removed,
			RequestRefused:     Removed,
			Removed:            Removed,
			Direct:             Removed,
		},
	},
	AdminCancelsInvitation: {
		UpdateFromType: map[GroupGroupType]bool{InvitationSent: true},
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
		UpdateFromType: map[GroupGroupType]bool{RequestSent: true},
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
}

type GroupGroupTransitionResults struct {
	Success   map[int64]bool `json:"success,omitempty"`
	Unchanged map[int64]bool `json:"unchanged,omitempty"`
	Invalid   map[int64]bool `json:"invalid,omitempty"`
	Cycle     map[int64]bool `json:"cycle,omitempty"`
}

func (s *GroupGroupStore) transition(action GroupGroupTransitionAction, parentGroupID int64, childGroupIDs []int64) *GroupGroupTransitionResults {
	s.mustBeInTransaction()

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

		if len(idsToDelete) > 0 {
			idsToDeleteSlice := make([]int64, 0, len(idsToDelete))
			for id := range idsToDelete {
				idsToDeleteSlice = append(idsToDeleteSlice, id)
			}
			mustNotBeError(s.Delete("idGroupParent = ? AND idGroupChild IN (?)", parentGroupID, idsToDeleteSlice).Error())
		}

		shouldCreateNewAncestors := false
		if len(idsToUpdate) > 0 {
			updateData := map[GroupGroupType][]int64{}
			for id, toType := range idsToUpdate {
				updateData[toType] = append(updateData[toType], id)
				if toType.IsActive() {
					shouldCreateNewAncestors = true
				}
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
					if toType.IsActive() {
						shouldCreateNewAncestors = true
					}
				}
				return s.db.Exec(insertQuery, values...).Error
			}))
		}

		if shouldCreateNewAncestors {
			s.createNewAncestors()
		}
		return nil
	}))
	return &results
}
