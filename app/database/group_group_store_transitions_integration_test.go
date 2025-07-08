//go:build !unit

package database_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
	_ "unsafe"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

type groupGroup struct {
	ParentGroupID              int64
	ChildGroupID               int64
	ExpiresAt                  string
	PersonalInfoViewApprovedAt *database.Time
	LockMembershipApprovedAt   *database.Time
	WatchApprovedAt            *database.Time
}

type groupPendingRequest struct {
	GroupID                  int64
	MemberID                 int64
	Type                     string
	PersonalInfoViewApproved bool
	LockMembershipApproved   bool
	WatchApproved            bool
}

type groupMembershipChange struct {
	GroupID     int64
	MemberID    int64
	Action      string
	InitiatorID *int64
	At          *database.Time
}

type groupAncestor struct {
	AncestorGroupID int64
	ChildGroupID    int64
	IsSelf          bool
	ExpiresAt       string
}

type transitionTest struct {
	name                        string
	action                      database.GroupGroupTransitionAction
	relationsToChange           []int64
	approvals                   map[int64]database.GroupApprovals
	createPendingCycleWithType  string
	doNotEnforceMaxParticipants bool
	maxParticipants             *int
	wantResult                  database.GroupGroupTransitionResults
	wantApprovalsToRequest      map[int64]database.GroupApprovals
	wantGroupGroups             []groupGroup
	wantGroupPendingRequests    []groupPendingRequest
	wantGroupAncestors          []groupAncestor
	wantGroupMembershipChanges  []groupMembershipChange
	wantGrantedPermissions      []grantedPermission
	wantGeneratedPermissions    []permissionsGeneratedResultRow
	shouldRunListeners          bool
}

var (
	allTheIDs                  = []int64{1, 2, 3, 4, 5, 6, 7, 10, 11, 20, 30}
	allPossibleGroupsAncestors = []groupAncestor{
		{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
		{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
		{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
		{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
		{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
		{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
		{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
		{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
		{AncestorGroupID: 11, ChildGroupID: 11, IsSelf: true},
		{AncestorGroupID: 20, ChildGroupID: 1},
		{AncestorGroupID: 20, ChildGroupID: 2},
		{AncestorGroupID: 20, ChildGroupID: 3},
		{AncestorGroupID: 20, ChildGroupID: 4},
		{AncestorGroupID: 20, ChildGroupID: 5},
		{AncestorGroupID: 20, ChildGroupID: 6},
		{AncestorGroupID: 20, ChildGroupID: 7},
		{AncestorGroupID: 20, ChildGroupID: 10},
		{AncestorGroupID: 20, ChildGroupID: 11},
		{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
		{AncestorGroupID: 20, ChildGroupID: 50},
		{AncestorGroupID: 20, ChildGroupID: 51},
		{AncestorGroupID: 20, ChildGroupID: 52},
		{AncestorGroupID: 20, ChildGroupID: 53},
		{AncestorGroupID: 20, ChildGroupID: 54},
		{AncestorGroupID: 20, ChildGroupID: 55},
		{AncestorGroupID: 30, ChildGroupID: 1},
		{AncestorGroupID: 30, ChildGroupID: 2},
		{AncestorGroupID: 30, ChildGroupID: 3},
		{AncestorGroupID: 30, ChildGroupID: 4},
		{AncestorGroupID: 30, ChildGroupID: 5},
		{AncestorGroupID: 30, ChildGroupID: 6},
		{AncestorGroupID: 30, ChildGroupID: 7},
		{AncestorGroupID: 30, ChildGroupID: 10},
		{AncestorGroupID: 30, ChildGroupID: 11},
		{AncestorGroupID: 30, ChildGroupID: 20},
		{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
		{AncestorGroupID: 30, ChildGroupID: 50},
		{AncestorGroupID: 30, ChildGroupID: 51},
		{AncestorGroupID: 30, ChildGroupID: 52},
		{AncestorGroupID: 30, ChildGroupID: 53},
		{AncestorGroupID: 30, ChildGroupID: 54},
		{AncestorGroupID: 30, ChildGroupID: 55},
		{AncestorGroupID: 50, ChildGroupID: 50, IsSelf: true},
		{AncestorGroupID: 51, ChildGroupID: 51, IsSelf: true},
		{AncestorGroupID: 52, ChildGroupID: 52, IsSelf: true},
		{AncestorGroupID: 53, ChildGroupID: 53, IsSelf: true},
		{AncestorGroupID: 54, ChildGroupID: 54, IsSelf: true},
		{AncestorGroupID: 55, ChildGroupID: 55, IsSelf: true},
		{AncestorGroupID: 111, ChildGroupID: 111, IsSelf: true},
	}
)

var groupAncestorsUnchanged = []groupAncestor{
	{AncestorGroupID: 1, ChildGroupID: 1, IsSelf: true},
	{AncestorGroupID: 2, ChildGroupID: 2, IsSelf: true},
	{AncestorGroupID: 3, ChildGroupID: 3, IsSelf: true},
	{AncestorGroupID: 4, ChildGroupID: 4, IsSelf: true},
	{AncestorGroupID: 5, ChildGroupID: 5, IsSelf: true},
	{AncestorGroupID: 6, ChildGroupID: 6, IsSelf: true},
	{AncestorGroupID: 7, ChildGroupID: 7, IsSelf: true},
	{AncestorGroupID: 10, ChildGroupID: 10, IsSelf: true},
	{AncestorGroupID: 11, ChildGroupID: 11, IsSelf: true},
	{AncestorGroupID: 20, ChildGroupID: 4},
	{AncestorGroupID: 20, ChildGroupID: 5},
	{AncestorGroupID: 20, ChildGroupID: 6, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 20, ChildGroupID: 7, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 20, ChildGroupID: 10},
	{AncestorGroupID: 20, ChildGroupID: 11},
	{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
	{AncestorGroupID: 20, ChildGroupID: 50},
	{AncestorGroupID: 20, ChildGroupID: 51},
	{AncestorGroupID: 20, ChildGroupID: 52},
	{AncestorGroupID: 20, ChildGroupID: 53},
	{AncestorGroupID: 20, ChildGroupID: 54},
	{AncestorGroupID: 20, ChildGroupID: 55},
	{AncestorGroupID: 30, ChildGroupID: 4},
	{AncestorGroupID: 30, ChildGroupID: 5},
	{AncestorGroupID: 30, ChildGroupID: 6, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 30, ChildGroupID: 7, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 30, ChildGroupID: 10},
	{AncestorGroupID: 30, ChildGroupID: 11},
	{AncestorGroupID: 30, ChildGroupID: 20},
	{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
	{AncestorGroupID: 30, ChildGroupID: 50},
	{AncestorGroupID: 30, ChildGroupID: 51},
	{AncestorGroupID: 30, ChildGroupID: 52},
	{AncestorGroupID: 30, ChildGroupID: 53},
	{AncestorGroupID: 30, ChildGroupID: 54},
	{AncestorGroupID: 30, ChildGroupID: 55},
	{AncestorGroupID: 50, ChildGroupID: 50, IsSelf: true},
	{AncestorGroupID: 51, ChildGroupID: 51, IsSelf: true},
	{AncestorGroupID: 52, ChildGroupID: 52, IsSelf: true},
	{AncestorGroupID: 53, ChildGroupID: 53, IsSelf: true},
	{AncestorGroupID: 54, ChildGroupID: 54, IsSelf: true},
	{AncestorGroupID: 55, ChildGroupID: 55, IsSelf: true},
	{AncestorGroupID: 111, ChildGroupID: 111, IsSelf: true},
}

var groupsGroupsUnchanged = []groupGroup{
	{ParentGroupID: 20, ChildGroupID: 4},
	{ParentGroupID: 20, ChildGroupID: 5},
	{ParentGroupID: 20, ChildGroupID: 6, ExpiresAt: "2019-05-30 11:00:00"},
	{ParentGroupID: 20, ChildGroupID: 7, ExpiresAt: "2019-05-30 11:00:00"},
	{ParentGroupID: 20, ChildGroupID: 10},
	{ParentGroupID: 20, ChildGroupID: 11},
	{ParentGroupID: 20, ChildGroupID: 50},
	{ParentGroupID: 20, ChildGroupID: 51},
	{ParentGroupID: 20, ChildGroupID: 52},
	{ParentGroupID: 20, ChildGroupID: 53},
	{ParentGroupID: 20, ChildGroupID: 54},
	{ParentGroupID: 20, ChildGroupID: 55},
	{ParentGroupID: 30, ChildGroupID: 20},
}

var groupPendingRequestsUnchanged = []groupPendingRequest{
	{GroupID: 20, MemberID: 2, Type: "invitation"},
	{GroupID: 20, MemberID: 3, Type: "join_request"},
	{GroupID: 20, MemberID: 5, Type: "leave_request"},
	{GroupID: 20, MemberID: 7, Type: "leave_request"},
}

var grantedPermissionsUnchanged = []grantedPermission{
	{GroupID: 4, ItemID: 1, SourceGroupID: 4, Origin: "group_membership", CanView: "info"},
	{GroupID: 4, ItemID: 1, SourceGroupID: 20, Origin: "group_membership", CanView: "solution"},
	{GroupID: 4, ItemID: 1, SourceGroupID: 20, Origin: "item_unlocking", CanView: "content"},
	{GroupID: 5, ItemID: 2, SourceGroupID: 5, Origin: "group_membership", CanView: "info"},
	{GroupID: 5, ItemID: 2, SourceGroupID: 20, Origin: "group_membership", CanView: "solution"},
	{GroupID: 5, ItemID: 2, SourceGroupID: 20, Origin: "item_unlocking", CanView: "content"},
	{GroupID: 10, ItemID: 1, SourceGroupID: 10, Origin: "group_membership", CanView: "info"},
	{GroupID: 10, ItemID: 1, SourceGroupID: 20, Origin: "group_membership", CanView: "solution"},
	{GroupID: 10, ItemID: 1, SourceGroupID: 20, Origin: "item_unlocking", CanView: "content"},
	{GroupID: 11, ItemID: 2, SourceGroupID: 11, Origin: "group_membership", CanView: "info"},
	{GroupID: 11, ItemID: 2, SourceGroupID: 20, Origin: "group_membership", CanView: "solution"},
	{GroupID: 11, ItemID: 2, SourceGroupID: 20, Origin: "item_unlocking", CanView: "content"},
}

var generatedPermissionsUnchanged = []permissionsGeneratedResultRow{
	{GroupID: 4, ItemID: 1, CanViewGenerated: "solution"},
	{GroupID: 5, ItemID: 2, CanViewGenerated: "solution"},
	{GroupID: 10, ItemID: 1, CanViewGenerated: "solution"},
	{GroupID: 11, ItemID: 2, CanViewGenerated: "solution"},
}

var (
	currentTimePtr = (*database.Time)(golang.Ptr(time.Now().UTC()))
	userID         = int64(111)
	userIDPtr      = &userID
)

const maxDateTime = "9999-12-31 23:59:59"

func testTransitionAcceptingNoRelationAndAnyPendingRequest(name string, action database.GroupGroupTransitionAction,
	expectedGroupMembershipAction database.GroupMembershipAction,
	doNotEnforceMaxParticipants bool, maxParticipants *int,
) transitionTest {
	resultForDirectRelations := database.Invalid
	return transitionTest{
		name:                        name,
		action:                      action,
		relationsToChange:           allTheIDs,
		doNotEnforceMaxParticipants: doNotEnforceMaxParticipants,
		maxParticipants:             maxParticipants,
		wantResult: database.GroupGroupTransitionResults{
			1: "success", 2: "success", 3: "success", 6: "success", 7: "success",

			4: resultForDirectRelations, 5: resultForDirectRelations, 10: resultForDirectRelations, 11: resultForDirectRelations,
			20: "invalid",
			30: "cycle",
		},
		wantGroupPendingRequests: []groupPendingRequest{
			{GroupID: 20, MemberID: 5, Type: "leave_request"},
		},
		wantGroupGroups: []groupGroup{
			{ParentGroupID: 20, ChildGroupID: 1},
			{ParentGroupID: 20, ChildGroupID: 2},
			{ParentGroupID: 20, ChildGroupID: 3},
			{ParentGroupID: 20, ChildGroupID: 6},
			{ParentGroupID: 20, ChildGroupID: 4},
			{ParentGroupID: 20, ChildGroupID: 5},
			{ParentGroupID: 20, ChildGroupID: 7},
			{ParentGroupID: 20, ChildGroupID: 10},
			{ParentGroupID: 20, ChildGroupID: 11},
			{ParentGroupID: 20, ChildGroupID: 50},
			{ParentGroupID: 20, ChildGroupID: 51},
			{ParentGroupID: 20, ChildGroupID: 52},
			{ParentGroupID: 20, ChildGroupID: 53},
			{ParentGroupID: 20, ChildGroupID: 54},
			{ParentGroupID: 20, ChildGroupID: 55},
			{ParentGroupID: 30, ChildGroupID: 20},
		},
		wantGroupAncestors: allPossibleGroupsAncestors,
		wantGroupMembershipChanges: []groupMembershipChange{
			{GroupID: 20, MemberID: 1, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 2, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 3, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 6, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 7, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
		},
		wantGrantedPermissions: grantedPermissionsUnchanged,
		shouldRunListeners:     true,
	}
}

func testTransitionAcceptingNoRelationAndAnyPendingRequestEnforcingMaxParticipants(name string, action database.GroupGroupTransitionAction,
	acceptDirectRelations bool,
) transitionTest {
	resultForDirectRelations := database.Invalid
	if acceptDirectRelations {
		resultForDirectRelations = database.Unchanged
	}
	return transitionTest{
		name:              name,
		action:            action,
		relationsToChange: allTheIDs,
		maxParticipants:   golang.Ptr(8),
		wantResult: database.GroupGroupTransitionResults{
			1: "full", 2: "full", 3: "full", 6: "full", 7: "full",

			4: resultForDirectRelations, 5: resultForDirectRelations, 10: resultForDirectRelations, 11: resultForDirectRelations,
			20: "invalid",
			30: "cycle",
		},
		wantGroupPendingRequests:   groupPendingRequestsUnchanged,
		wantGroupGroups:            groupsGroupsUnchanged,
		wantGroupAncestors:         groupAncestorsUnchanged,
		wantGroupMembershipChanges: nil,
		wantGrantedPermissions:     grantedPermissionsUnchanged,
		shouldRunListeners:         false,
	}
}

func testTransitionAcceptingPendingRequest(name string, action database.GroupGroupTransitionAction,
	acceptedID int64, pendingType, expectedGroupMembershipAction database.GroupMembershipAction,
	doNotEnforceMaxParticipants bool, maxParticipants *int,
) transitionTest {
	return transitionTest{
		name:                        name,
		action:                      action,
		relationsToChange:           allTheIDs,
		createPendingCycleWithType:  pendingType.PendingType(),
		doNotEnforceMaxParticipants: doNotEnforceMaxParticipants,
		maxParticipants:             maxParticipants,
		wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
			acceptedID: "success", 30: "cycle",
		}),
		wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, nil,
			[]groupGroup{{ParentGroupID: 20, ChildGroupID: acceptedID}}),
		wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, pendingType.PendingType(),
			map[string]*groupPendingRequest{fmt.Sprintf("20_%d", acceptedID): nil}, nil),
		wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
			nil,
			[]groupAncestor{
				{AncestorGroupID: 20, ChildGroupID: acceptedID},
				{AncestorGroupID: 30, ChildGroupID: acceptedID},
			}),
		wantGroupMembershipChanges: []groupMembershipChange{
			{GroupID: 20, MemberID: acceptedID, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
		},
		wantGrantedPermissions: grantedPermissionsUnchanged,
		shouldRunListeners:     true,
	}
}

func testTransitionAcceptingPendingRequestEnforcingMaxParticipants(name string, action database.GroupGroupTransitionAction,
	acceptedID int64, pendingType database.GroupMembershipAction,
) transitionTest {
	return transitionTest{
		name:                       name + " (enforcing max participants)",
		action:                     action,
		relationsToChange:          allTheIDs,
		createPendingCycleWithType: pendingType.PendingType(),
		maxParticipants:            golang.Ptr(5),
		wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
			acceptedID: "full", 30: "cycle",
		}),
		wantGroupGroups:            groupsGroupsUnchanged,
		wantGroupPendingRequests:   patchGroupPendingRequests(groupPendingRequestsUnchanged, pendingType.PendingType(), nil, nil),
		wantGroupAncestors:         groupAncestorsUnchanged,
		wantGroupMembershipChanges: nil,
		wantGrantedPermissions:     grantedPermissionsUnchanged,
		shouldRunListeners:         false,
	}
}

func testTransitionRemovingUserFromGroup(name string, action database.GroupGroupTransitionAction,
	expectedGroupMembershipAction database.GroupMembershipAction,
) transitionTest {
	return transitionTest{
		name:              name,
		action:            action,
		relationsToChange: allTheIDs,
		wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
			4: "success", 5: "success", 10: "success", 11: "success",
		}),
		wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged,
			map[string]*groupGroup{"20_4": nil, "20_5": nil, "20_10": nil, "20_11": nil}, nil),
		wantGroupPendingRequests: []groupPendingRequest{
			{GroupID: 20, MemberID: 2, Type: "invitation"},
			{GroupID: 20, MemberID: 3, Type: "join_request"},
			{GroupID: 20, MemberID: 7, Type: "leave_request"},
		},
		wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
			map[string]*groupAncestor{
				"20_4": nil, "20_5": nil, "20_10": nil, "20_11": nil,
				"30_4": nil, "30_5": nil, "30_10": nil, "30_11": nil,
			}, nil),
		wantGroupMembershipChanges: []groupMembershipChange{
			{GroupID: 20, MemberID: 4, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 5, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 10, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
			{GroupID: 20, MemberID: 11, Action: string(expectedGroupMembershipAction), At: currentTimePtr, InitiatorID: userIDPtr},
		},
		wantGrantedPermissions: patchGrantedPermissions(grantedPermissionsUnchanged, []string{
			"4_1_20_group_membership", "5_2_20_group_membership", "10_1_20_group_membership", "11_2_20_group_membership",
		}),
		wantGeneratedPermissions: patchGeneratedPermissions(generatedPermissionsUnchanged, map[string]string{
			"4_1": "content", "5_2": "content", "10_1": "content", "11_2": "content",
		}),
		shouldRunListeners: true,
	}
}

func TestGroupGroupStore_Transition(t *testing.T) {
	tests := []transitionTest{
		{
			name:                        "AdminCreatesInvitation",
			action:                      database.AdminCreatesInvitation,
			createPendingCycleWithType:  "join_request",
			relationsToChange:           allTheIDs,
			doNotEnforceMaxParticipants: true,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 3: "success", 6: "success", 7: "success",
				2: "unchanged",
				4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged,
				nil, []groupGroup{{ParentGroupID: 20, ChildGroupID: 3}}),
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "join_request",
				map[string]*groupPendingRequest{"20_3": nil, "20_7": {GroupID: 20, MemberID: 7, Type: "invitation"}},
				[]groupPendingRequest{
					{GroupID: 20, MemberID: 1, Type: "invitation"},
					{GroupID: 20, MemberID: 6, Type: "invitation"},
				}),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged, nil,
				[]groupAncestor{
					{AncestorGroupID: 20, ChildGroupID: 3},
					{AncestorGroupID: 30, ChildGroupID: 3},
				}),
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 1, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 3, Action: "join_request_accepted", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 6, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 7, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     true,
		},
		{
			name:                       "AdminCreatesInvitation (max participants limit is not exceeded)",
			action:                     database.AdminCreatesInvitation,
			createPendingCycleWithType: "join_request",
			relationsToChange:          allTheIDs,
			maxParticipants:            golang.Ptr(9),
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 3: "success", 6: "success", 7: "success",
				2: "unchanged",
				4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged,
				nil, []groupGroup{{ParentGroupID: 20, ChildGroupID: 3}}),
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "join_request",
				map[string]*groupPendingRequest{"20_3": nil, "20_7": {GroupID: 20, MemberID: 7, Type: "invitation"}},
				[]groupPendingRequest{
					{GroupID: 20, MemberID: 1, Type: "invitation"},
					{GroupID: 20, MemberID: 6, Type: "invitation"},
				}),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged, nil,
				[]groupAncestor{
					{AncestorGroupID: 20, ChildGroupID: 3},
					{AncestorGroupID: 30, ChildGroupID: 3},
				}),
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 1, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 3, Action: "join_request_accepted", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 6, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
				{GroupID: 20, MemberID: 7, Action: "invitation_created", InitiatorID: userIDPtr, At: currentTimePtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     true,
		},
		{
			name:                       "AdminCreatesInvitation (enforce max participants)",
			action:                     database.AdminCreatesInvitation,
			createPendingCycleWithType: "join_request",
			relationsToChange:          allTheIDs,
			maxParticipants:            golang.Ptr(8),
			wantResult: database.GroupGroupTransitionResults{
				1: "full", 3: "full", 6: "full", 7: "full",
				2: "unchanged",
				4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups:            groupsGroupsUnchanged,
			wantGroupPendingRequests:   patchGroupPendingRequests(groupPendingRequestsUnchanged, "join_request", nil, nil),
			wantGroupAncestors:         groupAncestorsUnchanged,
			wantGroupMembershipChanges: nil,
			wantGrantedPermissions:     grantedPermissionsUnchanged,
			shouldRunListeners:         false,
		},
		{
			name:                        "UserCreatesJoinRequest",
			action:                      database.UserCreatesJoinRequest,
			relationsToChange:           allTheIDs,
			doNotEnforceMaxParticipants: true,
			approvals: map[int64]database.GroupApprovals{
				1: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true},
				6: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: false},
				7: {PersonalInfoViewApproval: false, LockMembershipApproval: false, WatchApproval: true},
			},
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 6: "success", 7: "success",
				3: "unchanged",
				2: "invalid", 4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{
					"20_7": {GroupID: 20, MemberID: 7, Type: "join_request", WatchApproved: true},
				},
				[]groupPendingRequest{
					{
						GroupID: 20, MemberID: 1, Type: "join_request", PersonalInfoViewApproved: true,
						LockMembershipApproved: true, WatchApproved: true,
					},
					{
						GroupID: 20, MemberID: 6, Type: "join_request", PersonalInfoViewApproved: true,
						LockMembershipApproved: true,
					},
				}),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 1, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 6, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 7, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		{
			name:              "UserCreatesJoinRequest (max participants limit is not exceeded)",
			action:            database.UserCreatesJoinRequest,
			relationsToChange: allTheIDs,
			maxParticipants:   golang.Ptr(6),
			approvals: map[int64]database.GroupApprovals{
				1: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true},
				6: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: false},
				7: {PersonalInfoViewApproval: false, LockMembershipApproval: false, WatchApproval: true},
			},
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 6: "success", 7: "success",
				3: "unchanged",
				2: "invalid", 4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{
					"20_7": {GroupID: 20, MemberID: 7, Type: "join_request", WatchApproved: true},
				},
				[]groupPendingRequest{
					{
						GroupID: 20, MemberID: 1, Type: "join_request", PersonalInfoViewApproved: true,
						LockMembershipApproved: true, WatchApproved: true,
					},
					{
						GroupID: 20, MemberID: 6, Type: "join_request", PersonalInfoViewApproved: true,
						LockMembershipApproved: true,
					},
				}),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 1, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 6, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 7, Action: "join_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		{
			name:              "UserCreatesJoinRequest (enforce max participants)",
			action:            database.UserCreatesJoinRequest,
			relationsToChange: allTheIDs,
			approvals: map[int64]database.GroupApprovals{
				1: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true},
				6: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: false},
				7: {PersonalInfoViewApproval: false, LockMembershipApproval: false, WatchApproval: true},
			},
			maxParticipants: golang.Ptr(5),
			wantResult: database.GroupGroupTransitionResults{
				1: "full", 6: "full", 7: "full",
				3: "unchanged",
				2: "invalid", 4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups:            groupsGroupsUnchanged,
			wantGroupPendingRequests:   patchGroupPendingRequests(groupPendingRequestsUnchanged, "", nil, nil),
			wantGroupAncestors:         groupAncestorsUnchanged,
			wantGroupMembershipChanges: nil,
			wantGrantedPermissions:     grantedPermissionsUnchanged,
			shouldRunListeners:         false,
		},
		testTransitionAcceptingPendingRequest("UserAcceptsInvitation",
			database.UserAcceptsInvitation, 2, database.InvitationCreated, database.InvitationAccepted, false, nil),
		{
			name:                       "UserAcceptsInvitation (should not do anything when all transitions cause cycles)",
			action:                     database.UserAcceptsInvitation,
			relationsToChange:          []int64{30},
			createPendingCycleWithType: "invitation",
			wantResult:                 database.GroupGroupTransitionResults{30: "cycle"},
			wantGroupGroups:            groupsGroupsUnchanged,
			wantGroupPendingRequests:   patchGroupPendingRequests(groupPendingRequestsUnchanged, "invitation", nil, nil),
			wantGroupAncestors:         patchGroupAncestors(groupAncestorsUnchanged, nil, nil),
			wantGrantedPermissions:     grantedPermissionsUnchanged,
			shouldRunListeners:         false,
		},
		testTransitionAcceptingPendingRequest("AdminAcceptsJoinRequest",
			database.AdminAcceptsJoinRequest, 3, database.JoinRequestCreated, database.JoinRequestAccepted, true, nil),
		testTransitionAcceptingPendingRequest("AdminAcceptsJoinRequest (max participants limit is not exceeded)",
			database.AdminAcceptsJoinRequest, 3, database.JoinRequestCreated, database.JoinRequestAccepted, false, golang.Ptr(7)),
		testTransitionAcceptingPendingRequestEnforcingMaxParticipants(
			"AdminAcceptsJoinRequest", database.AdminAcceptsJoinRequest, 3, database.JoinRequestCreated),
		{
			name:              "UserRefusesInvitation",
			action:            database.UserRefusesInvitation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				2: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_2": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 2, Action: "invitation_refused", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		{
			name:              "AdminRefusesJoinRequest",
			action:            database.AdminRefusesJoinRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				3: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_3": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 3, Action: "join_request_refused", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		testTransitionRemovingUserFromGroup("AdminRemovesUser", database.AdminRemovesUser, database.Removed),
		{
			name:              "AdminWithdrawsInvitation",
			action:            database.AdminWithdrawsInvitation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				2: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_2": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 2, Action: "invitation_withdrawn", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		testTransitionRemovingUserFromGroup("UserLeavesGroup", database.UserLeavesGroup, database.Left),
		{
			name:              "UserCancelsJoinRequest",
			action:            database.UserCancelsJoinRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				3: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_3": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 3, Action: "join_request_withdrawn", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		{
			name:              "UserCreatesLeaveRequest",
			action:            database.UserCreatesLeaveRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				4: "success", 10: "success", 11: "success", 5: "unchanged",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				nil, []groupPendingRequest{
					{GroupID: 20, MemberID: 4, Type: "leave_request"},
					{GroupID: 20, MemberID: 10, Type: "leave_request"},
					{GroupID: 20, MemberID: 11, Type: "leave_request"},
				}),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 4, Action: "leave_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 10, Action: "leave_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
				{GroupID: 20, MemberID: 11, Action: "leave_request_created", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
		},
		{
			name:              "UserCancelsLeaveRequest",
			action:            database.UserCancelsLeaveRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				5: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_5": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 5, Action: "leave_request_withdrawn", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
		},
		{
			name:              "AdminAcceptsLeaveRequest",
			action:            database.AdminAcceptsLeaveRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				5: "success",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, map[string]*groupGroup{"20_5": nil}, nil),
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_5": nil}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged, map[string]*groupAncestor{"20_5": nil, "30_5": nil}, nil),
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 5, Action: "leave_request_accepted", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: patchGrantedPermissions(grantedPermissionsUnchanged, []string{
				"5_2_20_group_membership",
			}),
			wantGeneratedPermissions: patchGeneratedPermissions(generatedPermissionsUnchanged, map[string]string{
				"5_2": "content",
			}),
			shouldRunListeners: true,
		},
		{
			name:              "AdminRefusesLeaveRequest",
			action:            database.AdminRefusesLeaveRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				5: "success",
			}),
			wantGroupGroups: groupsGroupsUnchanged,
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_5": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			wantGroupMembershipChanges: []groupMembershipChange{
				{GroupID: 20, MemberID: 5, Action: "leave_request_refused", At: currentTimePtr, InitiatorID: userIDPtr},
			},
			wantGrantedPermissions: grantedPermissionsUnchanged,
			shouldRunListeners:     false,
		},
		testTransitionAcceptingNoRelationAndAnyPendingRequest(
			"UserCreatesAcceptedJoinRequest", database.UserCreatesAcceptedJoinRequest, database.JoinRequestAccepted, true, nil),
		testTransitionAcceptingNoRelationAndAnyPendingRequest("UserCreatesAcceptedJoinRequest (max participants limit is not exceeded)",
			database.UserCreatesAcceptedJoinRequest, database.JoinRequestAccepted, false, golang.Ptr(9)),
		testTransitionAcceptingNoRelationAndAnyPendingRequestEnforcingMaxParticipants(
			"UserCreatesAcceptedJoinRequest (enforce max participants)",
			database.UserCreatesAcceptedJoinRequest, false),
		testTransitionAcceptingNoRelationAndAnyPendingRequest(
			"UserJoinsGroupByBadge", database.UserJoinsGroupByBadge, database.JoinedByBadge, true, nil),
		testTransitionAcceptingNoRelationAndAnyPendingRequest("UserJoinsGroupByBadge (max participants limit is not exceeded)",
			database.UserJoinsGroupByBadge, database.JoinedByBadge, false, golang.Ptr(9)),
		testTransitionAcceptingNoRelationAndAnyPendingRequestEnforcingMaxParticipants(
			"UserJoinsGroupByBadge (enforce max participants)", database.UserJoinsGroupByBadge, false),
		testTransitionAcceptingNoRelationAndAnyPendingRequest(
			"UserJoinsGroupByCode", database.UserJoinsGroupByCode, database.JoinedByCode, true, nil),
		testTransitionAcceptingNoRelationAndAnyPendingRequest("UserJoinsGroupByCode (max participants limit is not exceeded)",
			database.UserJoinsGroupByCode, database.JoinedByCode, false, golang.Ptr(9)),
		testTransitionAcceptingNoRelationAndAnyPendingRequestEnforcingMaxParticipants(
			"UserJoinsGroupByCode (enforce max participants)", database.UserJoinsGroupByCode, false),
		{
			name:              "AdminRemovesDirectRelation",
			action:            database.AdminRemovesDirectRelation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				4: "success", 5: "success", 10: "success", 11: "success",
				1: "unchanged", 6: "unchanged", 30: "unchanged",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, map[string]*groupGroup{
				"20_4": nil, "20_5": nil, "20_10": nil, "20_11": nil,
			}, nil),
			wantGroupPendingRequests: patchGroupPendingRequests(groupPendingRequestsUnchanged, "",
				map[string]*groupPendingRequest{"20_5": nil}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				map[string]*groupAncestor{
					"20_4": nil, "20_5": nil, "20_10": nil, "20_11": nil,
					"30_4": nil, "30_5": nil, "30_10": nil, "30_11": nil,
				}, nil),
			wantGrantedPermissions: patchGrantedPermissions(grantedPermissionsUnchanged, []string{
				"4_1_20_group_membership", "5_2_20_group_membership", "10_1_20_group_membership", "11_2_20_group_membership",
			}),
			wantGeneratedPermissions: patchGeneratedPermissions(generatedPermissionsUnchanged, map[string]string{
				"4_1": "content", "5_2": "content", "10_1": "content", "11_2": "content",
			}),
			shouldRunListeners: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixture("group_group_store/transition/")
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			if tt.createPendingCycleWithType != "" {
				for _, id := range []int64{30, 20} {
					assert.NoError(t, dataStore.Exec(`
						INSERT INTO group_pending_requests (
							group_id, member_id, type, personal_info_view_approved, lock_membership_approved, watch_approved, at
						) VALUES (20, ?, ?, 1, 1, 1, NOW(3))`,
						id, tt.createPendingCycleWithType).Error())
				}
			}

			if tt.maxParticipants != nil {
				assert.NoError(t, dataStore.Groups().ByID(20).UpdateColumn("max_participants", *tt.maxParticipants).Error())
			}

			if tt.doNotEnforceMaxParticipants {
				assert.NoError(t, dataStore.Groups().ByID(20).UpdateColumn("enforce_max_participants", false).Error())
			}

			var result database.GroupGroupTransitionResults
			var approvalsToRequest map[int64]database.GroupApprovals
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, approvalsToRequest, err = store.GroupGroups().Transition(
					tt.action, 20, tt.relationsToChange, tt.approvals, userID,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			if tt.wantApprovalsToRequest == nil {
				tt.wantApprovalsToRequest = map[int64]database.GroupApprovals{}
			}
			assert.Equal(t, tt.wantApprovalsToRequest, approvalsToRequest)

			assertGroupGroupsEqual(t, dataStore.GroupGroups(), tt.wantGroupGroups)
			assertGroupPendingRequestsEqual(t, dataStore.GroupPendingRequests(), tt.wantGroupPendingRequests)

			assertGroupAncestorsEqual(t, dataStore, &tt)

			var count int64
			assert.NoError(t, dataStore.Table("groups_propagate").
				Where("ancestors_computation_state != 'done'").Count(&count).Error())
			if tt.shouldRunListeners {
				assert.Zero(t, count, "Listeners should be executed")
			} else {
				assert.NotZero(t, count, "Listeners should not be executed")
			}

			assertGroupMembershipChangesEqual(t, dataStore.GroupMembershipChanges(), tt.wantGroupMembershipChanges)
			assertGrantedPermissionsEqual(t, dataStore.PermissionsGranted(), tt.wantGrantedPermissions)
			assertGeneratedPermissionsEqual(t, dataStore.Permissions(), tt.wantGeneratedPermissions)
		})
	}
}

func assertGroupAncestorsEqual(t *testing.T, dataStore *database.DataStore, tt *transitionTest) {
	var groupAncestors []groupAncestor
	assert.NoError(t, dataStore.GroupAncestors().Select("ancestor_group_id, child_group_id, is_self, expires_at").
		Order("ancestor_group_id, child_group_id").Scan(&groupAncestors).Error())

	sort.Slice(tt.wantGroupAncestors, func(i, j int) bool {
		return tt.wantGroupAncestors[i].AncestorGroupID < tt.wantGroupAncestors[j].AncestorGroupID ||
			(tt.wantGroupAncestors[i].AncestorGroupID == tt.wantGroupAncestors[j].AncestorGroupID &&
				tt.wantGroupAncestors[i].ChildGroupID < tt.wantGroupAncestors[j].ChildGroupID)
	})

	for i := 0; i < len(tt.wantGroupAncestors); i++ {
		if tt.wantGroupAncestors[i].ExpiresAt == "" {
			tt.wantGroupAncestors[i].ExpiresAt = maxDateTime
		}
		if tt.shouldRunListeners {
			parsed, err := time.Parse(time.DateTime, tt.wantGroupAncestors[i].ExpiresAt)
			assert.NoError(t, err)
			if parsed.Before(time.Now().UTC()) {
				newValue := make([]groupAncestor, 0, len(tt.wantGroupAncestors)-1)
				newValue = append(newValue, tt.wantGroupAncestors[0:i]...)
				newValue = append(newValue, tt.wantGroupAncestors[i+1:len(tt.wantGroupAncestors)]...)
				tt.wantGroupAncestors = newValue
				i--
			}
		}
	}
	assert.Equal(t, tt.wantGroupAncestors, groupAncestors)
}

type approvalsTest struct {
	name                               string
	requirePersonalInfoAccessApproval  string
	requireLockMembershipApprovalUntil string
	requireWatchApproval               int
	personalInfoViewApproved           int
	lockMembershipApproved             int
	watchApproved                      int
	wantResult                         database.GroupGroupTransitionResult
	wantApprovalsToRequest             database.GroupApprovals
	wantPersonalInfoViewApprovedAt     *database.Time
	wantLockMembershipApprovedAt       *database.Time
	wantWatchApprovedAt                *database.Time
}

const success = "success"

func generateApprovalsTests(expectedTime *database.Time) []approvalsTest {
	return []approvalsTest{
		{
			name:                              "no approvals required, no approvals given",
			wantResult:                        success,
			requirePersonalInfoAccessApproval: "none",
		},
		{
			name:                              "no approvals required, personal_info_view approval given",
			wantResult:                        success,
			requirePersonalInfoAccessApproval: "none",
			personalInfoViewApproved:          1,
			wantPersonalInfoViewApprovedAt:    expectedTime,
		},
		{
			name:                              "no approvals required, lock_membership approval given",
			wantResult:                        success,
			requirePersonalInfoAccessApproval: "none",
			lockMembershipApproved:            1,
			wantLockMembershipApprovedAt:      expectedTime,
		},
		{
			name:                              "no approvals required, watch approval given",
			wantResult:                        success,
			requirePersonalInfoAccessApproval: "none",
			watchApproved:                     1,
			wantWatchApprovedAt:               expectedTime,
		},
		{
			name:                              "personal_info_view approval required (view), but it is not given",
			wantResult:                        "approvals_missing",
			requirePersonalInfoAccessApproval: "view",
			wantApprovalsToRequest:            database.GroupApprovals{PersonalInfoViewApproval: true},
		},
		{
			name:                              "personal_info_view approval required (edit), but it is not given",
			wantResult:                        "approvals_missing",
			requirePersonalInfoAccessApproval: "edit",
			wantApprovalsToRequest:            database.GroupApprovals{PersonalInfoViewApproval: true},
		},
		{
			name:                               "lock_membership_until is expired",
			wantResult:                         success,
			requirePersonalInfoAccessApproval:  "none",
			requireLockMembershipApprovalUntil: "2019-05-30 11:00:00",
		},
		{
			name:                               "lock_membership_until is not expired, but the lock_membership approval is not given",
			wantResult:                         "approvals_missing",
			requirePersonalInfoAccessApproval:  "none",
			requireLockMembershipApprovalUntil: "9999-12-31 23:59:59",
			wantApprovalsToRequest:             database.GroupApprovals{LockMembershipApproval: true},
		},
		{
			name:                              "watch approval required, but it is not given",
			wantResult:                        "approvals_missing",
			requirePersonalInfoAccessApproval: "none",
			requireWatchApproval:              1,
			wantApprovalsToRequest:            database.GroupApprovals{WatchApproval: true},
		},
		{
			name:                               "all approvals required, but personal_info_view is not given",
			wantResult:                         "approvals_missing",
			requirePersonalInfoAccessApproval:  "view",
			requireLockMembershipApprovalUntil: "9999-12-31 23:59:59",
			requireWatchApproval:               1,
			personalInfoViewApproved:           0,
			lockMembershipApproved:             1,
			watchApproved:                      1,
			wantApprovalsToRequest:             database.GroupApprovals{PersonalInfoViewApproval: true},
		},
		{
			name:                               "all approvals required, but lock_membership is not given",
			wantResult:                         "approvals_missing",
			requirePersonalInfoAccessApproval:  "view",
			requireLockMembershipApprovalUntil: "9999-12-31 23:59:59",
			requireWatchApproval:               1,
			personalInfoViewApproved:           1,
			lockMembershipApproved:             0,
			watchApproved:                      1,
			wantApprovalsToRequest:             database.GroupApprovals{LockMembershipApproval: true},
		},
		{
			name:                               "all approvals required, but watch is not given",
			wantResult:                         "approvals_missing",
			requirePersonalInfoAccessApproval:  "view",
			requireLockMembershipApprovalUntil: "9999-12-31 23:59:59",
			requireWatchApproval:               1,
			personalInfoViewApproved:           1,
			lockMembershipApproved:             1,
			watchApproved:                      0,
			wantApprovalsToRequest:             database.GroupApprovals{WatchApproval: true},
		},
	}
}

func TestGroupGroupStore_Transition_ChecksApprovalsInJoinRequestsOnAcceptingJoinRequests(t *testing.T) {
	expectedTime := (*database.Time)(golang.Ptr(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC)))
	for _, tt := range generateApprovalsTests(expectedTime) {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(fmt.Sprintf(`
				groups:
					- {id: 3}
					- {id: 20, require_personal_info_access_approval: %s, require_lock_membership_approval_until: %s, require_watch_approval: %d}
					- {id: 111}
				users:
					- {group_id: 111}
				groups_propagate: [{id: 3, ancestors_computation_state: todo}]
				group_pending_requests:
					- {group_id: 20, member_id: 3, type: join_request, at: 2019-05-30 11:00:00.001,
					   personal_info_view_approved: %d, lock_membership_approved: %d, watch_approved: %d}`,
				tt.requirePersonalInfoAccessApproval, tt.requireLockMembershipApprovalUntil, tt.requireWatchApproval,
				tt.personalInfoViewApproved, tt.lockMembershipApproved, tt.watchApproved))
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			var result database.GroupGroupTransitionResults
			var approvalsToRequest map[int64]database.GroupApprovals
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, approvalsToRequest, err = store.GroupGroups().Transition(
					database.AdminAcceptsJoinRequest, 20, []int64{3}, nil, 111,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, database.GroupGroupTransitionResults{3: tt.wantResult}, result)

			if tt.wantResult == success {
				assert.Empty(t, approvalsToRequest)
				assertGroupGroupsEqual(t, dataStore.GroupGroups(), []groupGroup{
					{
						ParentGroupID:              20,
						ChildGroupID:               3,
						PersonalInfoViewApprovedAt: tt.wantPersonalInfoViewApprovedAt,
						LockMembershipApprovedAt:   tt.wantLockMembershipApprovedAt,
						WatchApprovedAt:            tt.wantWatchApprovedAt,
					},
				})
				assertGroupPendingRequestsEqual(t, dataStore.GroupPendingRequests(), nil)
			} else {
				assert.Equal(t, map[int64]database.GroupApprovals{3: tt.wantApprovalsToRequest}, approvalsToRequest)
				assertGroupGroupsEqual(t, dataStore.GroupGroups(), nil)
			}

			var count int64
			assert.NoError(t, dataStore.Table("groups_propagate").
				Where("ancestors_computation_state != 'done'").Count(&count).Error())
			if tt.wantResult == success {
				assert.Zero(t, count, "Listeners should be executed")
			} else {
				assert.NotZero(t, count, "Listeners should not be executed")
			}
		})
	}
}

func TestGroupGroupStore_Transition_ChecksApprovalsInJoinRequestIfJoinRequestExists(t *testing.T) {
	for _, test := range []struct {
		name   string
		action database.GroupGroupTransitionAction
	}{
		{"when a user joins the group by code", database.UserJoinsGroupByCode},
		{"when a user joins the group by badge", database.UserJoinsGroupByBadge},
		{"when a group owner creates an accepted join request", database.UserCreatesAcceptedJoinRequest},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(`
				groups:
					- {id: 3}
					- {id: 20, require_personal_info_access_approval: view,
					   require_lock_membership_approval_until: 9999-12-31 23:59:59, require_watch_approval: 1}
					- {id: 111}
				users:
					- {group_id: 111}
				group_pending_requests:
					- {group_id: 20, member_id: 3, type: join_request, at: 2019-05-30 11:00:00.001,
					   personal_info_view_approved: 0, lock_membership_approved: 0, watch_approved: 0}`)
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			var result database.GroupGroupTransitionResults
			var approvalsToRequest map[int64]database.GroupApprovals
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, approvalsToRequest, err = store.GroupGroups().Transition(
					test.action, 20, []int64{3}, nil, 111,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, database.GroupGroupTransitionResults{3: "approvals_missing"}, result)
			assert.Equal(t, map[int64]database.GroupApprovals{
				3: {PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true},
			}, approvalsToRequest)
		})
	}
}

func TestGroupGroupStore_Transition_ReplacesJoinRequestByInvitationWhenNotNotEnoughApprovalsInJoinRequestOnCreatingInvitation(
	t *testing.T,
) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups:
			- {id: 3}
			- {id: 20, require_personal_info_access_approval: view,
				 require_lock_membership_approval_until: 9999-12-31 23:59:59, require_watch_approval: 1}
			- {id: 111}
		users:
			- {group_id: 111}
		group_pending_requests:
			- {group_id: 20, member_id: 3, type: join_request, at: 2019-05-30 11:00:00.001,
				 personal_info_view_approved: 0, lock_membership_approved: 0, watch_approved: 0}`)
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	var result database.GroupGroupTransitionResults
	var approvalsToRequest map[int64]database.GroupApprovals
	err := dataStore.InTransaction(func(store *database.DataStore) error {
		var err error
		result, approvalsToRequest, err = store.GroupGroups().Transition(
			database.AdminCreatesInvitation, 20, []int64{3}, nil, 111,
		)
		return err
	})

	assert.NoError(t, err)
	assert.Equal(t, database.GroupGroupTransitionResults{3: "success"}, result)
	assert.Empty(t, approvalsToRequest)
	assertGroupPendingRequestsEqual(t, dataStore.GroupPendingRequests(), []groupPendingRequest{
		{GroupID: 20, MemberID: 3, Type: "invitation"},
	})
}

func TestGroupGroupStore_Transition_ChecksApprovalsFromParametersOnAcceptingInvitations(t *testing.T) {
	const success = "success"
	expectedTime := (*database.Time)(golang.Ptr(time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)))
	database.MockNow("2019-06-01 00:00:00")
	defer database.RestoreNow()

	for _, tt := range generateApprovalsTests(expectedTime) {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := testhelpers.SetupDBWithFixtureString(fmt.Sprintf(`
				groups:
					- {id: 3}
					- {id: 20, require_personal_info_access_approval: %s, require_lock_membership_approval_until: %s, require_watch_approval: %d}
					- {id: 111}
				users:
					- {group_id: 111}
				groups_propagate: [{id: 3, ancestors_computation_state: todo}]
				group_pending_requests:
					- {group_id: 20, member_id: 3, type: invitation, at: 2019-05-30 11:00:00.001}`,
				tt.requirePersonalInfoAccessApproval, tt.requireLockMembershipApprovalUntil, tt.requireWatchApproval))
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			var result database.GroupGroupTransitionResults
			var approvalsToRequest map[int64]database.GroupApprovals
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, approvalsToRequest, err = store.GroupGroups().Transition(
					database.UserAcceptsInvitation, 20, []int64{3}, map[int64]database.GroupApprovals{
						3: {
							PersonalInfoViewApproval: tt.personalInfoViewApproved == 1,
							LockMembershipApproval:   tt.lockMembershipApproved == 1,
							WatchApproval:            tt.watchApproved == 1,
						},
					}, 111,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, database.GroupGroupTransitionResults{3: tt.wantResult}, result)

			if tt.wantResult == success {
				assert.Empty(t, approvalsToRequest)
				assertGroupGroupsEqual(t, dataStore.GroupGroups(), []groupGroup{
					{
						ParentGroupID:              20,
						ChildGroupID:               3,
						PersonalInfoViewApprovedAt: tt.wantPersonalInfoViewApprovedAt,
						LockMembershipApprovedAt:   tt.wantLockMembershipApprovedAt,
						WatchApprovedAt:            tt.wantWatchApprovedAt,
					},
				})
				assertGroupPendingRequestsEqual(t, dataStore.GroupPendingRequests(), nil)
			} else {
				assert.Equal(t, map[int64]database.GroupApprovals{3: tt.wantApprovalsToRequest}, approvalsToRequest)
				assertGroupGroupsEqual(t, dataStore.GroupGroups(), nil)
			}

			var count int64
			assert.NoError(t, dataStore.Table("groups_propagate").
				Where("ancestors_computation_state != 'done'").Count(&count).Error())
			if tt.wantResult == success {
				assert.Zero(t, count, "Listeners should be executed")
			} else {
				assert.NotZero(t, count, "Listeners should not be executed")
			}
		})
	}
}

func patchGroupGroups(old []groupGroup, diff map[string]*groupGroup,
	added []groupGroup,
) []groupGroup {
	result := make([]groupGroup, 0, len(old)+len(added))
	for _, relation := range old {
		if patch, ok := diff[fmt.Sprintf("%d_%d", relation.ParentGroupID, relation.ChildGroupID)]; ok {
			if patch == nil {
				continue // the relation is deleted
			}
			relation = *patch
		}
		result = append(result, relation)
	}
	result = append(result, added...)
	return result
}

func patchGroupPendingRequests(old []groupPendingRequest, cycleWithType string, diff map[string]*groupPendingRequest,
	added []groupPendingRequest,
) []groupPendingRequest {
	result := make([]groupPendingRequest, 0, len(old)+len(added))
	for _, relation := range old {
		if patch, ok := diff[fmt.Sprintf("%d_%d", relation.GroupID, relation.MemberID)]; ok {
			if patch == nil {
				continue // the relation is deleted
			}
			relation = *patch
		}
		result = append(result, relation)
	}
	result = append(result, added...)
	if cycleWithType != "" {
		result = append(result,
			groupPendingRequest{
				GroupID: 20, MemberID: 20, Type: cycleWithType, PersonalInfoViewApproved: true,
				LockMembershipApproved: true, WatchApproved: true,
			},
			groupPendingRequest{
				GroupID: 20, MemberID: 30, Type: cycleWithType, PersonalInfoViewApproved: true,
				LockMembershipApproved: true, WatchApproved: true,
			},
		)
	}
	return result
}

func patchGroupAncestors(old []groupAncestor, diff map[string]*groupAncestor, added []groupAncestor) []groupAncestor {
	result := make([]groupAncestor, 0, len(old)+len(added))
	for _, relation := range old {
		if patch, ok := diff[fmt.Sprintf("%d_%d", relation.AncestorGroupID, relation.ChildGroupID)]; ok {
			if patch == nil {
				continue // the relation is deleted
			}
			relation = *patch
		}
		result = append(result, relation)
	}
	result = append(result, added...)
	return result
}

func patchGrantedPermissions(old []grantedPermission, deleteIDs []string) []grantedPermission {
	patchMap := make(map[string]bool, len(deleteIDs))
	for _, deleteID := range deleteIDs {
		patchMap[deleteID] = true
	}
	result := make([]grantedPermission, 0, len(old)-len(patchMap))
	for _, permission := range old {
		if !patchMap[fmt.Sprintf("%d_%d_%d_%s", permission.GroupID, permission.ItemID, permission.SourceGroupID, permission.Origin)] {
			result = append(result, permission)
		}
	}
	return result
}

func patchGeneratedPermissions(
	old []permissionsGeneratedResultRow, canViewGeneratedChangeMap map[string]string,
) []permissionsGeneratedResultRow {
	result := make([]permissionsGeneratedResultRow, 0, len(old))
	for _, permission := range old {
		result = append(result, permission)
		if newCanViewGenerated, ok := canViewGeneratedChangeMap[fmt.Sprintf("%d_%d", permission.GroupID, permission.ItemID)]; ok {
			result[len(result)-1].CanViewGenerated = newCanViewGenerated
		}
	}
	return result
}

func buildExpectedGroupTransitionResults(nonInvalid database.GroupGroupTransitionResults) database.GroupGroupTransitionResults {
	result := make(database.GroupGroupTransitionResults, 12)
	const invalid = "invalid"
	for i := int64(1); i <= 11; i++ {
		if i < 8 || i > 9 {
			result[i] = invalid
		}
	}
	result[20] = invalid
	result[30] = invalid
	for key, value := range nonInvalid {
		result[key] = value
	}
	return result
}

func assertGroupGroupsEqual(t *testing.T, groupGroupStore *database.GroupGroupStore, expected []groupGroup) {
	var groupsGroups []groupGroup
	assert.NoError(t, groupGroupStore.Select(`
			parent_group_id, child_group_id, expires_at, personal_info_view_approved_at,
			lock_membership_approved_at, watch_approved_at`).
		Order("parent_group_id, child_group_id").Scan(&groupsGroups).Error())

	assert.Len(t, groupsGroups, len(expected))
	if len(groupsGroups) != len(expected) {
		return
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].ParentGroupID < expected[j].ParentGroupID ||
			expected[i].ParentGroupID == expected[j].ParentGroupID && expected[i].ChildGroupID < expected[j].ChildGroupID
	})
	for index, row := range expected {
		assert.Equal(t, row.ParentGroupID, groupsGroups[index].ParentGroupID, "wrong parent group id for row %#v", groupsGroups[index])
		assert.Equal(t, row.ChildGroupID, groupsGroups[index].ChildGroupID, "wrong child group id for row %#v", groupsGroups[index])
		if row.ExpiresAt == "" {
			row.ExpiresAt = maxDateTime
		}
		assert.Equal(t, row.ExpiresAt, groupsGroups[index].ExpiresAt, "wrong expires_at for row %#v", groupsGroups[index])
		assert.Equal(t, row.PersonalInfoViewApprovedAt, groupsGroups[index].PersonalInfoViewApprovedAt,
			"wrong personal_info_view_approved_at for row %#v", groupsGroups[index])
		assert.Equal(t, row.LockMembershipApprovedAt, groupsGroups[index].LockMembershipApprovedAt,
			"wrong lock_membership_approved_at for row %#v", groupsGroups[index])
		assert.Equal(t, row.WatchApprovedAt, groupsGroups[index].WatchApprovedAt,
			"wrong watch_approved_at for row %#v", groupsGroups[index])
	}
}

func assertGroupPendingRequestsEqual(t *testing.T, groupPendingRequestStore *database.GroupPendingRequestStore,
	expected []groupPendingRequest,
) {
	var groupPendingRequests []groupPendingRequest
	assert.NoError(t, groupPendingRequestStore.Select(`
			group_id, member_id, `+"`type`"+`, personal_info_view_approved,
			lock_membership_approved, watch_approved`).
		Order("group_id, member_id").Scan(&groupPendingRequests).Error())

	assert.Len(t, groupPendingRequests, len(expected))
	if len(groupPendingRequests) != len(expected) {
		return
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].GroupID < expected[j].GroupID ||
			expected[i].GroupID == expected[j].GroupID && expected[i].MemberID < expected[j].MemberID
	})
	for index, row := range expected {
		assert.Equal(t, row.GroupID, groupPendingRequests[index].GroupID, "wrong group id for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.MemberID, groupPendingRequests[index].MemberID, "wrong member id for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.Type, groupPendingRequests[index].Type, "wrong type for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.PersonalInfoViewApproved, groupPendingRequests[index].PersonalInfoViewApproved,
			"wrong personal_info_view_approved for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.LockMembershipApproved, groupPendingRequests[index].LockMembershipApproved,
			"wrong lock_membership_approved for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.WatchApproved, groupPendingRequests[index].WatchApproved,
			"wrong lock_membership_approved for row %#v", groupPendingRequests[index])
	}
}

func assertGroupMembershipChangesEqual(
	t *testing.T, groupMembershipChangeStore *database.GroupMembershipChangeStore, expected []groupMembershipChange,
) {
	var groupMembershipChanges []groupMembershipChange
	assert.NoError(t, groupMembershipChangeStore.Select("group_id, member_id, initiator_id, action, at").
		Order("group_id, member_id, at").Scan(&groupMembershipChanges).Error())

	assert.Len(t, groupMembershipChanges, len(expected))
	if len(groupMembershipChanges) != len(expected) {
		return
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].GroupID < expected[j].GroupID ||
			expected[i].GroupID == expected[j].GroupID && expected[i].MemberID < expected[j].MemberID
	})
	for index, row := range expected {
		assert.Equal(t, row.GroupID, groupMembershipChanges[index].GroupID, "group id for row %#v", groupMembershipChanges[index])
		assert.Equal(t, row.MemberID, groupMembershipChanges[index].MemberID, "wrong member id for row %#v", groupMembershipChanges[index])
		assert.Equal(t, row.Action, groupMembershipChanges[index].Action, "wrong action for row %#v", groupMembershipChanges[index])
		assert.Equal(t, row.InitiatorID, groupMembershipChanges[index].InitiatorID,
			"wrong initiator_id for row %#v", groupMembershipChanges[index])
		if groupMembershipChanges[index].At != nil {
			assert.True(t, (*time.Time)(groupMembershipChanges[index].At).Sub(time.Now().UTC())/time.Second < 5)
			assert.True(t, time.Now().UTC().Sub(time.Time(*groupMembershipChanges[index].At))/time.Second > -5)
		}
	}
}

func assertGrantedPermissionsEqual(t *testing.T, grantedPermissionStore *database.PermissionGrantedStore, expected []grantedPermission) {
	var grantedPermissions []grantedPermission
	assert.NoError(t, grantedPermissionStore.Select("group_id, item_id, source_group_id, origin, can_view").
		Order("group_id, item_id, source_group_id, origin").Scan(&grantedPermissions).Error())
	assert.Equal(t, expected, grantedPermissions)
}

func assertGeneratedPermissionsEqual(
	t *testing.T, permissionGeneratedStore *database.PermissionGeneratedStore, expected []permissionsGeneratedResultRow,
) {
	if expected == nil {
		expected = make([]permissionsGeneratedResultRow, 0)
	}
	var generatedPermissions []permissionsGeneratedResultRow
	assert.NoError(t, permissionGeneratedStore.Select("group_id, item_id, can_view_generated").
		Order("group_id, item_id").Scan(&generatedPermissions).Error())
	assert.EqualValues(t, expected, generatedPermissions)
}

func Test_insertGroupMembershipChanges_Duplicate(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(`
		groups: [{id: 1}, {id: 2}, {id: 3}]
		users: [{group_id: 3}]
		group_membership_changes: [{group_id: 1, member_id: 2, action: join_request_created, initiator_id: 3, at: 2019-05-30 11:00:00.123}]`)
	defer func() { _ = db.Close() }()

	timeMs := 122
	callsCount := 0
	var patchGuard *monkey.PatchGuard
	patchGuard = monkey.PatchInstanceMethod(reflect.TypeOf(&gorm.DB{}), "Exec", func(db *gorm.DB, sql string, values ...interface{}) *gorm.DB {
		patchGuard.Unpatch()
		defer patchGuard.Restore()
		timeMs++
		callsCount++
		testhelpers.MockDBTime(fmt.Sprintf("2019-05-30 11:00:00.%03d", timeMs))
		defer testhelpers.RestoreDBTime()
		return db.Exec(sql, values...)
	})
	defer patchGuard.Unpatch()

	dataStore := database.NewDataStore(db)
	insertGroupMembershipChanges(dataStore, map[int64]database.GroupMembershipAction{2: database.JoinRequestCreated}, 1, 3)
	var count int64
	require.NoError(t, dataStore.Table("group_membership_changes").Count(&count).Error())
	assert.Equal(t, int64(2), count)

	assert.Equal(t, 2, callsCount)
}

//go:linkname insertGroupMembershipChanges github.com/France-ioi/AlgoreaBackend/v2/app/database.insertGroupMembershipChanges
func insertGroupMembershipChanges(*database.DataStore, map[int64]database.GroupMembershipAction, int64, int64)
