// +build !unit

package database_test

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type groupGroup struct {
	ParentGroupID int64      `gorm:"column:idGroupParent"`
	ChildGroupID  int64      `gorm:"column:idGroupChild"`
	Type          string     `gorm:"column:sType"`
	ChildOrder    int64      `gorm:"column:iChildOrder"`
	StatusDate    *time.Time `gorm:"column:sStatusDate"`
}

type groupAncestor struct {
	AncestorGroupID int64 `gorm:"column:idGroupAncestor"`
	ChildGroupID    int64 `gorm:"column:idGroupChild"`
	IsSelf          bool  `gorm:"column:bIsSelf"`
}

func TestGroupGroupStore_transition(t *testing.T) {
	currentTimePtr := ptrTime(time.Now().UTC())
	groupAncestorsUnchanged := []groupAncestor{
		{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
		{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
		{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
		{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
		{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
		{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
		{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
		{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
		{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
		{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
		{AncestorGroupID: 20, ChildGroupID: 4},
		{AncestorGroupID: 20, ChildGroupID: 5},
		{AncestorGroupID: 20, ChildGroupID: 10},
		{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
		{AncestorGroupID: 30, ChildGroupID: 4},
		{AncestorGroupID: 30, ChildGroupID: 5},
		{AncestorGroupID: 30, ChildGroupID: 10},
		{AncestorGroupID: 30, ChildGroupID: 20},
		{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
	}

	tests := []struct {
		name                string
		action              database.GroupGroupTransitionAction
		createCycleWithType database.GroupGroupType
		wantResult          *database.GroupGroupTransitionResults
		wantGroupGroups     []groupGroup
		wantGroupAncestors  []groupAncestor
	}{
		{
			name:                "AdminCreatesInvitation",
			action:              database.AdminCreatesInvitation,
			createCycleWithType: database.RequestSent,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{1: true, 3: true, 6: true, 7: true, 8: true, 9: true},
				Unchanged: map[int64]bool{2: true},
				Invalid:   map[int64]bool{4: true, 5: true, 10: true, 20: true},
				Cycle:     map[int64]bool{30: true},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 1, Type: "invitationSent", ChildOrder: 1, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestAccepted", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationSent", ChildOrder: 2, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "invitationSent", ChildOrder: 3, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "invitationSent", ChildOrder: 4, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "invitationSent", ChildOrder: 5, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 20, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 30, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 3},
				{AncestorGroupID: 20, ChildGroupID: 4},
				{AncestorGroupID: 20, ChildGroupID: 5},
				{AncestorGroupID: 20, ChildGroupID: 10},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 3},
				{AncestorGroupID: 30, ChildGroupID: 4},
				{AncestorGroupID: 30, ChildGroupID: 5},
				{AncestorGroupID: 30, ChildGroupID: 10},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
		{
			name:   "UserCreatesRequest",
			action: database.UserCreatesRequest,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{1: true, 6: true, 7: true, 8: true, 9: true},
				Unchanged: map[int64]bool{3: true},
				Invalid:   map[int64]bool{2: true, 4: true, 5: true, 10: true, 20: true},
				Cycle:     map[int64]bool{30: true},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 1, Type: "requestSent", ChildOrder: 1, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "requestSent", ChildOrder: 2, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestSent", ChildOrder: 3, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "requestSent", ChildOrder: 4, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "requestSent", ChildOrder: 5, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: groupAncestorsUnchanged,
		},
		{
			name:                "UserAcceptsInvitation",
			action:              database.UserAcceptsInvitation,
			createCycleWithType: database.InvitationSent,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{2: true},
				Unchanged: map[int64]bool{4: true},
				Invalid:   map[int64]bool{1: true, 3: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 20: true},
				Cycle:     map[int64]bool{30: true},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationAccepted", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 20, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 30, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 2},
				{AncestorGroupID: 20, ChildGroupID: 4},
				{AncestorGroupID: 20, ChildGroupID: 5},
				{AncestorGroupID: 20, ChildGroupID: 10},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 2},
				{AncestorGroupID: 30, ChildGroupID: 4},
				{AncestorGroupID: 30, ChildGroupID: 5},
				{AncestorGroupID: 30, ChildGroupID: 10},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
		{
			name:                "AdminAcceptsRequest",
			action:              database.AdminAcceptsRequest,
			createCycleWithType: database.RequestSent,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{3: true},
				Unchanged: map[int64]bool{5: true},
				Invalid:   map[int64]bool{1: true, 2: true, 4: true, 6: true, 7: true, 8: true, 9: true, 10: true, 20: true},
				Cycle:     map[int64]bool{30: true},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestAccepted", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 20, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 30, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 3},
				{AncestorGroupID: 20, ChildGroupID: 4},
				{AncestorGroupID: 20, ChildGroupID: 5},
				{AncestorGroupID: 20, ChildGroupID: 10},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 3},
				{AncestorGroupID: 30, ChildGroupID: 4},
				{AncestorGroupID: 30, ChildGroupID: 5},
				{AncestorGroupID: 30, ChildGroupID: 10},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
		{
			name:   "UserRefusesInvitation",
			action: database.UserRefusesInvitation,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{2: true},
				Unchanged: map[int64]bool{6: true},
				Invalid:   map[int64]bool{1: true, 3: true, 4: true, 5: true, 7: true, 8: true, 9: true, 10: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationRefused", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: groupAncestorsUnchanged,
		},
		{
			name:   "AdminRefusesRequest",
			action: database.AdminRefusesRequest,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{3: true},
				Unchanged: map[int64]bool{7: true},
				Invalid:   map[int64]bool{1: true, 2: true, 4: true, 5: true, 6: true, 8: true, 9: true, 10: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestRefused", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: groupAncestorsUnchanged,
		},
		{
			name:   "AdminRemovesChild",
			action: database.AdminRemovesChild,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{4: true, 5: true, 6: true, 7: true, 10: true},
				Unchanged: map[int64]bool{8: true},
				Invalid:   map[int64]bool{1: true, 2: true, 3: true, 9: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "removed", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "removed", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "removed", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "removed", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "removed", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
		{
			name:   "AdminCancelsInvitation",
			action: database.AdminCancelsInvitation,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{2: true},
				Unchanged: map[int64]bool{},
				Invalid:   map[int64]bool{1: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: groupAncestorsUnchanged,
		},
		{
			name:   "UserLeavesGroup",
			action: database.UserLeavesGroup,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{4: true, 5: true, 10: true},
				Unchanged: map[int64]bool{9: true},
				Invalid:   map[int64]bool{1: true, 2: true, 3: true, 6: true, 7: true, 8: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "left", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "left", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "left", ChildOrder: 0, StatusDate: currentTimePtr},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
		{
			name:   "UserCancelsRequest",
			action: database.UserCancelsRequest,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{3: true},
				Unchanged: map[int64]bool{},
				Invalid:   map[int64]bool{1: true, 2: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 20: true, 30: true},
				Cycle:     map[int64]bool{},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "removed", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "left", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 0, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: groupAncestorsUnchanged,
		},
		{
			name:   "AdminAddsDirectRelation",
			action: database.AdminAddsDirectRelation,
			wantResult: &database.GroupGroupTransitionResults{
				Success:   map[int64]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true},
				Unchanged: map[int64]bool{10: true},
				Invalid:   map[int64]bool{20: true},
				Cycle:     map[int64]bool{30: true},
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 1, Type: "direct", ChildOrder: 1, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 2, Type: "direct", ChildOrder: 2, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "direct", ChildOrder: 3, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "direct", ChildOrder: 4, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "direct", ChildOrder: 5, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "direct", ChildOrder: 6, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "direct", ChildOrder: 7, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "direct", ChildOrder: 8, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "direct", ChildOrder: 9, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 10, StatusDate: nil},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct", ChildOrder: 0, StatusDate: nil},
			},
			wantGroupAncestors: []groupAncestor{
				{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
				{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
				{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
				{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
				{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
				{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
				{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
				{AncestorGroupID: 8, ChildGroupID: 8, IsSelf: true},
				{AncestorGroupID: 9, ChildGroupID: 9, IsSelf: true},
				{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
				{AncestorGroupID: 20, ChildGroupID: 1},
				{AncestorGroupID: 20, ChildGroupID: 2},
				{AncestorGroupID: 20, ChildGroupID: 3},
				{AncestorGroupID: 20, ChildGroupID: 4},
				{AncestorGroupID: 20, ChildGroupID: 5},
				{AncestorGroupID: 20, ChildGroupID: 6},
				{AncestorGroupID: 20, ChildGroupID: 7},
				{AncestorGroupID: 20, ChildGroupID: 8},
				{AncestorGroupID: 20, ChildGroupID: 9},
				{AncestorGroupID: 20, ChildGroupID: 10},
				{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
				{AncestorGroupID: 30, ChildGroupID: 1},
				{AncestorGroupID: 30, ChildGroupID: 2},
				{AncestorGroupID: 30, ChildGroupID: 3},
				{AncestorGroupID: 30, ChildGroupID: 4},
				{AncestorGroupID: 30, ChildGroupID: 5},
				{AncestorGroupID: 30, ChildGroupID: 6},
				{AncestorGroupID: 30, ChildGroupID: 7},
				{AncestorGroupID: 30, ChildGroupID: 8},
				{AncestorGroupID: 30, ChildGroupID: 9},
				{AncestorGroupID: 30, ChildGroupID: 10},
				{AncestorGroupID: 30, ChildGroupID: 20},
				{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("group_group_store/transition/")
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			if tt.createCycleWithType != database.NoRelation {
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO groups_groups (idGroupParent, idGroupChild, sType) VALUES (20, 30, ?)", tt.createCycleWithType).Error())
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO groups_groups (idGroupParent, idGroupChild, sType) VALUES (20, 20, ?)", tt.createCycleWithType).Error())
			}
			var result *database.GroupGroupTransitionResults
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				result = store.GroupGroups().Transition(
					tt.action, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30},
				)
				return nil
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)

			assertGroupGroupsEqual(t, dataStore.GroupGroups(), tt.wantGroupGroups)

			var groupAncestors []groupAncestor
			assert.NoError(t, dataStore.GroupAncestors().Select("idGroupAncestor, idGroupChild, bIsSelf").
				Order("idGroupAncestor, idGroupChild").Scan(&groupAncestors).Error())
			assert.Equal(t, tt.wantGroupAncestors, groupAncestors)
		})
	}
}

func assertGroupGroupsEqual(t *testing.T, groupGroupStore *database.GroupGroupStore, expected []groupGroup) {
	var groupsGroups []groupGroup
	assert.NoError(t, groupGroupStore.Select("idGroupParent, idGroupChild, iChildOrder, sType, sStatusDate").
		Order("idGroupParent, idGroupChild").Scan(&groupsGroups).Error())

	assert.Len(t, groupsGroups, len(expected))
	if len(groupsGroups) != len(expected) {
		return
	}
	usedChildOrders := make(map[int64]bool, len(expected))
	for index, row := range expected {
		assert.Equal(t, row.ParentGroupID, groupsGroups[index].ParentGroupID, "wrong parent group ID for row %#v", groupsGroups[index])
		assert.Equal(t, row.ChildGroupID, groupsGroups[index].ChildGroupID, "wrong child group ID for row %#v", groupsGroups[index])
		assert.Equal(t, row.Type, groupsGroups[index].Type, "wrong type for row %#v", groupsGroups[index])
		if row.ChildOrder == 0 {
			assert.Zero(t, groupsGroups[index].ChildOrder)
		} else {
			assert.False(t, usedChildOrders[groupsGroups[index].ChildOrder])
			usedChildOrders[groupsGroups[index].ChildOrder] = true
		}
		if row.StatusDate == nil {
			assert.Nil(t, groupsGroups[index].StatusDate)
		} else {
			assert.NotNil(t, groupsGroups[index].StatusDate, "StatusDate should be nil in row %#v", groupsGroups[index])
			if groupsGroups[index].StatusDate != nil {
				assert.True(t, groupsGroups[index].StatusDate.Sub(time.Now().UTC())/time.Second < 5)
				assert.True(t, time.Now().UTC().Sub(*groupsGroups[index].StatusDate)/time.Second > -5)
			}
		}
	}
}
