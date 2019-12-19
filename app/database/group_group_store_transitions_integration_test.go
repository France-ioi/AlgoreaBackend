// +build !unit

package database_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

type groupGroup struct {
	ParentGroupID int64
	ChildGroupID  int64
	ChildOrder    int64
	ExpiresAt     string
}

type groupPendingRequest struct {
	GroupID                  int64
	MemberID                 int64
	Type                     string
	PersonalInfoViewApproved bool
	PersonalInfoEditApproved bool
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
	name                       string
	action                     database.GroupGroupTransitionAction
	relationsToChange          []int64
	approvals                  map[int64]database.GroupApprovals
	createPendingCycleWithType string
	wantResult                 database.GroupGroupTransitionResults
	wantGroupGroups            []groupGroup
	wantGroupPendingRequests   []groupPendingRequest
	wantGroupAncestors         []groupAncestor
	wantGroupMembershipChanges []groupMembershipChange
	shouldRunListeners         bool
}

var allTheIDs = []int64{1, 2, 3, 4, 5, 6, 7, 10, 11, 20, 30}
var allPossibleGroupsAncestors = []groupAncestor{
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
	{AncestorGroupID: 111, ChildGroupID: 111, IsSelf: true},
}

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
	{AncestorGroupID: 30, ChildGroupID: 4},
	{AncestorGroupID: 30, ChildGroupID: 5},
	{AncestorGroupID: 30, ChildGroupID: 6, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 30, ChildGroupID: 7, ExpiresAt: "2019-05-30 11:00:00"},
	{AncestorGroupID: 30, ChildGroupID: 10},
	{AncestorGroupID: 30, ChildGroupID: 11},
	{AncestorGroupID: 30, ChildGroupID: 20},
	{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
	{AncestorGroupID: 111, ChildGroupID: 111, IsSelf: true},
}

var groupsGroupsUnchanged = []groupGroup{
	{ParentGroupID: 20, ChildGroupID: 4},
	{ParentGroupID: 20, ChildGroupID: 5},
	{ParentGroupID: 20, ChildGroupID: 6, ExpiresAt: "2019-05-30 11:00:00"},
	{ParentGroupID: 20, ChildGroupID: 7, ExpiresAt: "2019-05-30 11:00:00"},
	{ParentGroupID: 20, ChildGroupID: 10},
	{ParentGroupID: 20, ChildGroupID: 11},
	{ParentGroupID: 30, ChildGroupID: 20},
}

var groupPendingRequestsUnchanged = []groupPendingRequest{
	{GroupID: 20, MemberID: 2, Type: "invitation"},
	{GroupID: 20, MemberID: 3, Type: "join_request"},
	{GroupID: 20, MemberID: 5, Type: "leave_request"},
	{GroupID: 20, MemberID: 7, Type: "leave_request"},
}

var currentTimePtr = (*database.Time)(ptrTime(time.Now().UTC()))
var userID = int64(111)
var userIDPtr = &userID

const maxDateTime = "9999-12-31 23:59:59"

func testTransitionAcceptingNoRelationAndAnyPendingRequest(name string, action database.GroupGroupTransitionAction,
	expectedGroupMembershipAction database.GroupMembershipAction, acceptDirectRelations bool) transitionTest {
	resultForDirectRelations := database.Invalid
	if acceptDirectRelations {
		resultForDirectRelations = database.Unchanged
	}
	return transitionTest{
		name:              name,
		action:            action,
		relationsToChange: allTheIDs,
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
			{ParentGroupID: 20, ChildGroupID: 1, ChildOrder: 1},
			{ParentGroupID: 20, ChildGroupID: 2, ChildOrder: 2},
			{ParentGroupID: 20, ChildGroupID: 3, ChildOrder: 3},
			{ParentGroupID: 20, ChildGroupID: 6, ChildOrder: 4},
			{ParentGroupID: 20, ChildGroupID: 4},
			{ParentGroupID: 20, ChildGroupID: 5},
			{ParentGroupID: 20, ChildGroupID: 7},
			{ParentGroupID: 20, ChildGroupID: 10},
			{ParentGroupID: 20, ChildGroupID: 11},
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
		shouldRunListeners: true,
	}
}

func testTransitionAcceptingPendingRequest(name string, action database.GroupGroupTransitionAction,
	acceptedID int64, pendingType, expectedGroupMembershipAction database.GroupMembershipAction) transitionTest {
	return transitionTest{
		name:                       name,
		action:                     action,
		relationsToChange:          allTheIDs,
		createPendingCycleWithType: pendingType.PendingType(),
		wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
			acceptedID: "success", 30: "cycle",
		}),
		wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, nil,
			[]groupGroup{{ParentGroupID: 20, ChildGroupID: acceptedID, ChildOrder: 1}}),
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
		shouldRunListeners: true,
	}
}

func testTransitionRemovingUserFromGroup(name string, action database.GroupGroupTransitionAction,
	expectedGroupMembershipAction database.GroupMembershipAction) transitionTest {
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
		shouldRunListeners: true,
	}
}

func TestGroupGroupStore_Transition(t *testing.T) {
	tests := []transitionTest{
		{
			name:                       "AdminCreatesInvitation",
			action:                     database.AdminCreatesInvitation,
			createPendingCycleWithType: "join_request",
			relationsToChange:          allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 3: "success", 6: "success", 7: "success",
				2: "unchanged",
				4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged,
				nil, []groupGroup{{ParentGroupID: 20, ChildGroupID: 3, ChildOrder: 1}}),
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
			shouldRunListeners: true,
		},
		{
			name:              "UserCreatesJoinRequest",
			action:            database.UserCreatesJoinRequest,
			relationsToChange: allTheIDs,
			approvals: map[int64]database.GroupApprovals{
				1: {PersonalInfoViewApproval: true, PersonalInfoEditApproval: true, LockMembershipApproval: true, WatchApproval: true},
				6: {PersonalInfoViewApproval: true, PersonalInfoEditApproval: false, LockMembershipApproval: true, WatchApproval: false},
				7: {PersonalInfoViewApproval: false, PersonalInfoEditApproval: true, LockMembershipApproval: false, WatchApproval: true},
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
					"20_7": {GroupID: 20, MemberID: 7, Type: "join_request", PersonalInfoEditApproved: true, WatchApproved: true},
				},
				[]groupPendingRequest{
					{
						GroupID: 20, MemberID: 1, Type: "join_request", PersonalInfoViewApproved: true,
						PersonalInfoEditApproved: true, LockMembershipApproved: true, WatchApproved: true,
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
			shouldRunListeners: false,
		},
		testTransitionAcceptingPendingRequest(
			"UserAcceptsInvitation", database.UserAcceptsInvitation, 2, database.InvitationCreated, database.InvitationAccepted),
		{
			name:                       "UserAcceptsInvitation (should not do anything when all transitions cause cycles)",
			action:                     database.UserAcceptsInvitation,
			relationsToChange:          []int64{30},
			createPendingCycleWithType: "invitation",
			wantResult:                 database.GroupGroupTransitionResults{30: "cycle"},
			wantGroupGroups:            groupsGroupsUnchanged,
			wantGroupPendingRequests:   patchGroupPendingRequests(groupPendingRequestsUnchanged, "invitation", nil, nil),
			wantGroupAncestors:         patchGroupAncestors(groupAncestorsUnchanged, nil, nil),
			shouldRunListeners:         false,
		},
		testTransitionAcceptingPendingRequest(
			"AdminAcceptsJoinRequest", database.AdminAcceptsJoinRequest, 3, database.JoinRequestCreated, database.JoinRequestAccepted),
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
			shouldRunListeners: false,
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
			shouldRunListeners: false,
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
			shouldRunListeners: false,
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
			shouldRunListeners: false,
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
			shouldRunListeners: false,
		},
		testTransitionAcceptingNoRelationAndAnyPendingRequest(
			"UserCreatesAcceptedJoinRequest", database.UserCreatesAcceptedJoinRequest, database.JoinRequestAccepted, false),
		testTransitionAcceptingNoRelationAndAnyPendingRequest(
			"UserJoinsGroupByCode", database.UserJoinsGroupByCode, database.JoinedByCode, false),
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
			shouldRunListeners: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("group_group_store/transition/")
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			if tt.createPendingCycleWithType != "" {
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO group_pending_requests (group_id, member_id, type) VALUES (20, 30, ?)", tt.createPendingCycleWithType).Error())
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO group_pending_requests (group_id, member_id, type) VALUES (20, 20, ?)", tt.createPendingCycleWithType).Error())
			}

			var result database.GroupGroupTransitionResults
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, err = store.GroupGroups().Transition(
					tt.action, 20, tt.relationsToChange, tt.approvals, userID,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)

			assertGroupGroupsEqual(t, dataStore.GroupGroups(), tt.wantGroupGroups)
			assertGroupPendingRequestsEqual(t, dataStore.GroupPendingRequests(), tt.wantGroupPendingRequests)

			var groupAncestors []groupAncestor
			assert.NoError(t, dataStore.GroupAncestors().Select("ancestor_group_id, child_group_id, is_self, expires_at").
				Order("ancestor_group_id, child_group_id").Scan(&groupAncestors).Error())

			sort.Slice(tt.wantGroupAncestors, func(i, j int) bool {
				return tt.wantGroupAncestors[i].AncestorGroupID < tt.wantGroupAncestors[j].AncestorGroupID ||
					(tt.wantGroupAncestors[i].AncestorGroupID == tt.wantGroupAncestors[j].AncestorGroupID &&
						tt.wantGroupAncestors[i].ChildGroupID < tt.wantGroupAncestors[j].ChildGroupID)
			})

			for index := range tt.wantGroupAncestors {
				if tt.wantGroupAncestors[index].ExpiresAt == "" {
					tt.wantGroupAncestors[index].ExpiresAt = maxDateTime
				}
			}
			assert.Equal(t, tt.wantGroupAncestors, groupAncestors)

			var count int64
			assert.NoError(t, dataStore.Table("groups_propagate").
				Where("ancestors_computation_state != 'done'").Count(&count).Error())
			if tt.shouldRunListeners {
				assert.Zero(t, count, "Listeners should be executed")
			} else {
				assert.NotZero(t, count, "Listeners should not be executed")
			}

			assertGroupMembershipChangesEqual(t, dataStore.GroupMembershipChanges(), tt.wantGroupMembershipChanges)
		})
	}
}

func patchGroupGroups(old []groupGroup, diff map[string]*groupGroup,
	added []groupGroup) []groupGroup {
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
	added []groupPendingRequest) []groupPendingRequest {
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
			groupPendingRequest{GroupID: 20, MemberID: 20, Type: cycleWithType},
			groupPendingRequest{GroupID: 20, MemberID: 30, Type: cycleWithType},
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
	assert.NoError(t, groupGroupStore.Select("parent_group_id, child_group_id, child_order, expires_at").
		Order("parent_group_id, child_group_id").Scan(&groupsGroups).Error())

	assert.Len(t, groupsGroups, len(expected))
	if len(groupsGroups) != len(expected) {
		return
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].ParentGroupID < expected[j].ParentGroupID ||
			expected[i].ParentGroupID == expected[j].ParentGroupID && expected[i].ChildGroupID < expected[j].ChildGroupID
	})
	usedChildOrders := make(map[int64]bool, len(expected))
	for index, row := range expected {
		assert.Equal(t, row.ParentGroupID, groupsGroups[index].ParentGroupID, "wrong parent group id for row %#v", groupsGroups[index])
		assert.Equal(t, row.ChildGroupID, groupsGroups[index].ChildGroupID, "wrong child group id for row %#v", groupsGroups[index])
		if row.ChildOrder == 0 {
			assert.Zero(t, groupsGroups[index].ChildOrder)
		} else {
			assert.False(t, usedChildOrders[groupsGroups[index].ChildOrder])
			usedChildOrders[groupsGroups[index].ChildOrder] = true
		}
		if row.ExpiresAt == "" {
			row.ExpiresAt = maxDateTime
		}
		assert.Equal(t, row.ExpiresAt, groupsGroups[index].ExpiresAt, "wrong expires_at for row %#v", groupsGroups[index])
	}
}

func assertGroupPendingRequestsEqual(t *testing.T, groupPendingRequestStore *database.GroupPendingRequestStore,
	expected []groupPendingRequest) {
	var groupPendingRequests []groupPendingRequest
	assert.NoError(t, groupPendingRequestStore.Select(`
			group_id, member_id, `+"`type`"+`, personal_info_view_approved, personal_info_edit_approved,
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
		assert.Equal(t, row.PersonalInfoEditApproved, groupPendingRequests[index].PersonalInfoEditApproved,
			"wrong personal_info_edit_approved for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.LockMembershipApproved, groupPendingRequests[index].LockMembershipApproved,
			"wrong lock_membership_approved for row %#v", groupPendingRequests[index])
		assert.Equal(t, row.WatchApproved, groupPendingRequests[index].WatchApproved,
			"wrong lock_membership_approved for row %#v", groupPendingRequests[index])
	}
}

func assertGroupMembershipChangesEqual(
	t *testing.T, groupMembershipChangeStore *database.GroupMembershipChangeStore, expected []groupMembershipChange) {
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
