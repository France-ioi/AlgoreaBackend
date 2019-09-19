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
	ParentGroupID  int64 `gorm:"column:group_parent_id"`
	ChildGroupID   int64 `gorm:"column:group_child_id"`
	Type           string
	InvitingUserID *int64 `gorm:"column:user_inviting_id"`
	ChildOrder     int64
	StatusDate     *database.Time
}

type groupAncestor struct {
	AncestorGroupID int64 `gorm:"column:group_ancestor_id"`
	ChildGroupID    int64 `gorm:"column:group_child_id"`
	IsSelf          bool
}

func TestGroupGroupStore_Transition(t *testing.T) {
	currentTimePtr := (*database.Time)(ptrTime(time.Now().UTC()))
	userID := int64(12)
	userIDPtr := &userID
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
		{AncestorGroupID: 11, ChildGroupID: 11, IsSelf: true},
		{AncestorGroupID: 20, ChildGroupID: 4},
		{AncestorGroupID: 20, ChildGroupID: 5},
		{AncestorGroupID: 20, ChildGroupID: 10},
		{AncestorGroupID: 20, ChildGroupID: 11},
		{AncestorGroupID: 20, ChildGroupID: 20, IsSelf: true},
		{AncestorGroupID: 30, ChildGroupID: 4},
		{AncestorGroupID: 30, ChildGroupID: 5},
		{AncestorGroupID: 30, ChildGroupID: 10},
		{AncestorGroupID: 30, ChildGroupID: 11},
		{AncestorGroupID: 30, ChildGroupID: 20},
		{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
	}
	groupsGroupsUnchanged := []groupGroup{
		{ParentGroupID: 20, ChildGroupID: 2, Type: "invitationSent"},
		{ParentGroupID: 20, ChildGroupID: 3, Type: "requestSent"},
		{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted"},
		{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted"},
		{ParentGroupID: 20, ChildGroupID: 6, Type: "invitationRefused"},
		{ParentGroupID: 20, ChildGroupID: 7, Type: "requestRefused"},
		{ParentGroupID: 20, ChildGroupID: 8, Type: "removed"},
		{ParentGroupID: 20, ChildGroupID: 9, Type: "left"},
		{ParentGroupID: 20, ChildGroupID: 10, Type: "direct"},
		{ParentGroupID: 20, ChildGroupID: 11, Type: "joinedByCode"},
		{ParentGroupID: 30, ChildGroupID: 20, Type: "direct"},
	}

	allTheIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 20, 30}
	allPossibleGroupsAncestors := []groupAncestor{
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
		{AncestorGroupID: 11, ChildGroupID: 11, IsSelf: true},
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
		{AncestorGroupID: 20, ChildGroupID: 11},
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
		{AncestorGroupID: 30, ChildGroupID: 11},
		{AncestorGroupID: 30, ChildGroupID: 20},
		{AncestorGroupID: 30, ChildGroupID: 30, IsSelf: true},
	}

	tests := []struct {
		name                string
		action              database.GroupGroupTransitionAction
		relationsToChange   []int64
		createCycleWithType database.GroupGroupType
		wantResult          database.GroupGroupTransitionResults
		wantGroupGroups     []groupGroup
		wantGroupAncestors  []groupAncestor
		shouldRunListeners  bool
	}{
		{
			name:                "AdminCreatesInvitation",
			action:              database.AdminCreatesInvitation,
			createCycleWithType: database.RequestSent,
			relationsToChange:   allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 3: "success", 6: "success", 7: "success", 8: "success", 9: "success",
				2: "unchanged",
				4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, database.RequestSent,
				map[string]*groupGroup{
					"20_3": {ParentGroupID: 20, ChildGroupID: 3, Type: "requestAccepted",
						InvitingUserID: userIDPtr, ChildOrder: 0, StatusDate: currentTimePtr},
					"20_6": {ParentGroupID: 20, ChildGroupID: 6, Type: "invitationSent",
						InvitingUserID: userIDPtr, ChildOrder: 2, StatusDate: currentTimePtr},
					"20_7": {ParentGroupID: 20, ChildGroupID: 7, Type: "invitationSent",
						InvitingUserID: userIDPtr, ChildOrder: 3, StatusDate: currentTimePtr},
					"20_8": {ParentGroupID: 20, ChildGroupID: 8, Type: "invitationSent",
						InvitingUserID: userIDPtr, ChildOrder: 4, StatusDate: currentTimePtr},
					"20_9": {ParentGroupID: 20, ChildGroupID: 9, Type: "invitationSent",
						InvitingUserID: userIDPtr, ChildOrder: 5, StatusDate: currentTimePtr},
				},
				[]groupGroup{
					{ParentGroupID: 20, ChildGroupID: 1, Type: "invitationSent", InvitingUserID: userIDPtr, ChildOrder: 1, StatusDate: currentTimePtr},
				}),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged, nil,
				[]groupAncestor{
					{AncestorGroupID: 20, ChildGroupID: 3},
					{AncestorGroupID: 30, ChildGroupID: 3},
				}),
			shouldRunListeners: true,
		},
		{
			name:              "UserCreatesRequest",
			action:            database.UserCreatesRequest,
			relationsToChange: allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 6: "success", 7: "success", 8: "success", 9: "success",
				3: "unchanged",
				2: "invalid", 4: "invalid", 5: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{
					"20_6": {ParentGroupID: 20, ChildGroupID: 6, Type: "requestSent", ChildOrder: 2, StatusDate: currentTimePtr},
					"20_7": {ParentGroupID: 20, ChildGroupID: 7, Type: "requestSent", ChildOrder: 3, StatusDate: currentTimePtr},
					"20_8": {ParentGroupID: 20, ChildGroupID: 8, Type: "requestSent", ChildOrder: 4, StatusDate: currentTimePtr},
					"20_9": {ParentGroupID: 20, ChildGroupID: 9, Type: "requestSent", ChildOrder: 5, StatusDate: currentTimePtr},
				},
				[]groupGroup{
					{ParentGroupID: 20, ChildGroupID: 1, Type: "requestSent", ChildOrder: 1, StatusDate: currentTimePtr},
				}),
			wantGroupAncestors: groupAncestorsUnchanged,
			shouldRunListeners: true,
		},
		{
			name:                "UserAcceptsInvitation",
			action:              database.UserAcceptsInvitation,
			relationsToChange:   allTheIDs,
			createCycleWithType: database.InvitationSent,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				2: "success", 4: "unchanged", 30: "cycle",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, database.InvitationSent,
				map[string]*groupGroup{
					"20_2": {ParentGroupID: 20, ChildGroupID: 2, Type: "invitationAccepted", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				nil,
				[]groupAncestor{
					{AncestorGroupID: 20, ChildGroupID: 2},
					{AncestorGroupID: 30, ChildGroupID: 2},
				}),
			shouldRunListeners: true,
		},
		{
			name:                "UserAcceptsInvitation (should not do anything when all transitions cause cycles)",
			action:              database.UserAcceptsInvitation,
			relationsToChange:   []int64{30},
			createCycleWithType: database.InvitationSent,
			wantResult:          database.GroupGroupTransitionResults{30: "cycle"},
			wantGroupGroups:     patchGroupGroups(groupsGroupsUnchanged, database.InvitationSent, nil, nil),
			wantGroupAncestors:  patchGroupAncestors(groupAncestorsUnchanged, nil, nil),
			shouldRunListeners:  false,
		},
		{
			name:                "AdminAcceptsRequest",
			action:              database.AdminAcceptsRequest,
			relationsToChange:   allTheIDs,
			createCycleWithType: database.RequestSent,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				3: "success", 5: "unchanged", 30: "cycle",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, database.RequestSent,
				map[string]*groupGroup{
					"20_3": {ParentGroupID: 20, ChildGroupID: 3, Type: "requestAccepted", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				nil,
				[]groupAncestor{
					{AncestorGroupID: 20, ChildGroupID: 3},
					{AncestorGroupID: 30, ChildGroupID: 3},
				}),
			shouldRunListeners: true,
		},
		{
			name:              "UserRefusesInvitation",
			action:            database.UserRefusesInvitation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				2: "success", 6: "unchanged",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{
					"20_2": {ParentGroupID: 20, ChildGroupID: 2, Type: "invitationRefused", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			shouldRunListeners: true,
		},
		{
			name:              "AdminRefusesRequest",
			action:            database.AdminRefusesRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				3: "success", 7: "unchanged",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{
					"20_3": {ParentGroupID: 20, ChildGroupID: 3, Type: "requestRefused", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			shouldRunListeners: true,
		},
		{
			name:              "AdminRemovesUser",
			action:            database.AdminRemovesUser,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				4: "success", 5: "success", 8: "unchanged", 11: "success",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{
					"20_4":  {ParentGroupID: 20, ChildGroupID: 4, Type: "removed", StatusDate: currentTimePtr},
					"20_5":  {ParentGroupID: 20, ChildGroupID: 5, Type: "removed", StatusDate: currentTimePtr},
					"20_11": {ParentGroupID: 20, ChildGroupID: 11, Type: "removed", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				map[string]*groupAncestor{"20_4": nil, "20_5": nil, "20_11": nil, "30_4": nil, "30_5": nil, "30_11": nil}, nil),
			shouldRunListeners: true,
		},
		{
			name:              "AdminCancelsInvitation",
			action:            database.AdminCancelsInvitation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				2: "success",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{"20_2": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			shouldRunListeners: true,
		},
		{
			name:              "UserLeavesGroup",
			action:            database.UserLeavesGroup,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				4: "success", 5: "success", 10: "success", 11: "success",
				9: "unchanged",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{
					"20_4":  {ParentGroupID: 20, ChildGroupID: 4, Type: "left", StatusDate: currentTimePtr},
					"20_5":  {ParentGroupID: 20, ChildGroupID: 5, Type: "left", StatusDate: currentTimePtr},
					"20_10": {ParentGroupID: 20, ChildGroupID: 10, Type: "left", StatusDate: currentTimePtr},
					"20_11": {ParentGroupID: 20, ChildGroupID: 11, Type: "left", StatusDate: currentTimePtr},
				}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				map[string]*groupAncestor{
					"20_4": nil, "20_5": nil, "20_10": nil, "20_11": nil,
					"30_4": nil, "30_5": nil, "30_10": nil, "30_11": nil,
				}, nil),
			shouldRunListeners: true,
		},
		{
			name:              "UserCancelsRequest",
			action:            database.UserCancelsRequest,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				3: "success",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{"20_3": nil}, nil),
			wantGroupAncestors: groupAncestorsUnchanged,
			shouldRunListeners: true,
		},
		{
			name:              "AdminAddsDirectRelation",
			action:            database.AdminAddsDirectRelation,
			relationsToChange: allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 2: "success", 3: "success", 4: "success", 5: "success", 6: "success", 7: "success", 8: "success",
				9: "success", 11: "success",

				10: "unchanged",
				20: "invalid",
				30: "cycle",
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
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct", ChildOrder: 10},
				{ParentGroupID: 20, ChildGroupID: 11, Type: "direct", ChildOrder: 11, StatusDate: currentTimePtr},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct"},
			},
			wantGroupAncestors: allPossibleGroupsAncestors,
			shouldRunListeners: true,
		},
		{
			name:              "UserCreatesAcceptedRequest",
			action:            database.UserCreatesAcceptedRequest,
			relationsToChange: allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 2: "success", 3: "success", 6: "success", 7: "success", 8: "success", 9: "success",
				5: "unchanged",
				4: "invalid", 10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "", map[string]*groupGroup{
				"20_2": {ParentGroupID: 20, ChildGroupID: 2, Type: "requestAccepted", ChildOrder: 2, StatusDate: currentTimePtr},
				"20_3": {ParentGroupID: 20, ChildGroupID: 3, Type: "requestAccepted", ChildOrder: 3, StatusDate: currentTimePtr},
				"20_6": {ParentGroupID: 20, ChildGroupID: 6, Type: "requestAccepted", ChildOrder: 4, StatusDate: currentTimePtr},
				"20_7": {ParentGroupID: 20, ChildGroupID: 7, Type: "requestAccepted", ChildOrder: 5, StatusDate: currentTimePtr},
				"20_8": {ParentGroupID: 20, ChildGroupID: 8, Type: "requestAccepted", ChildOrder: 6, StatusDate: currentTimePtr},
				"20_9": {ParentGroupID: 20, ChildGroupID: 9, Type: "requestAccepted", ChildOrder: 7, StatusDate: currentTimePtr},
			}, []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 1, Type: "requestAccepted", ChildOrder: 1, StatusDate: currentTimePtr},
			}),
			wantGroupAncestors: allPossibleGroupsAncestors,
			shouldRunListeners: true,
		},
		{
			name:              "UserJoinsGroupByCode",
			action:            database.UserJoinsGroupByCode,
			relationsToChange: allTheIDs,
			wantResult: database.GroupGroupTransitionResults{
				1: "success", 2: "success", 3: "success", 4: "invalid", 5: "invalid", 6: "success", 7: "success", 8: "success",
				9: "success",

				10: "invalid", 11: "invalid", 20: "invalid",
				30: "cycle",
			},
			wantGroupGroups: []groupGroup{
				{ParentGroupID: 20, ChildGroupID: 1, Type: "joinedByCode", ChildOrder: 1, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 2, Type: "joinedByCode", ChildOrder: 2, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 3, Type: "joinedByCode", ChildOrder: 3, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 4, Type: "invitationAccepted"},
				{ParentGroupID: 20, ChildGroupID: 5, Type: "requestAccepted"},
				{ParentGroupID: 20, ChildGroupID: 6, Type: "joinedByCode", ChildOrder: 4, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 7, Type: "joinedByCode", ChildOrder: 5, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 8, Type: "joinedByCode", ChildOrder: 6, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 9, Type: "joinedByCode", ChildOrder: 7, StatusDate: currentTimePtr},
				{ParentGroupID: 20, ChildGroupID: 10, Type: "direct"},
				{ParentGroupID: 20, ChildGroupID: 11, Type: "joinedByCode"},
				{ParentGroupID: 30, ChildGroupID: 20, Type: "direct"},
			},
			wantGroupAncestors: allPossibleGroupsAncestors,
			shouldRunListeners: true,
		},
		{
			name:              "AdminRemovesDirectRelation",
			action:            database.AdminRemovesDirectRelation,
			relationsToChange: allTheIDs,
			wantResult: buildExpectedGroupTransitionResults(database.GroupGroupTransitionResults{
				10: "success", 1: "unchanged", 30: "unchanged",
			}),
			wantGroupGroups: patchGroupGroups(groupsGroupsUnchanged, "",
				map[string]*groupGroup{"20_10": nil}, nil),
			wantGroupAncestors: patchGroupAncestors(groupAncestorsUnchanged,
				map[string]*groupAncestor{"20_10": nil, "30_10": nil}, nil),
			shouldRunListeners: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.SetupDBWithFixture("group_group_store/transition/")
			defer func() { _ = db.Close() }()
			dataStore := database.NewDataStore(db)

			if tt.createCycleWithType != database.NoRelation {
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO groups_groups (group_parent_id, group_child_id, type) VALUES (20, 30, ?)", tt.createCycleWithType).Error())
				assert.NoError(t, dataStore.Exec(
					"INSERT INTO groups_groups (group_parent_id, group_child_id, type) VALUES (20, 20, ?)", tt.createCycleWithType).Error())
			}
			var result database.GroupGroupTransitionResults
			err := dataStore.InTransaction(func(store *database.DataStore) error {
				var err error
				result, err = store.GroupGroups().Transition(
					tt.action, 20, tt.relationsToChange, 12,
				)
				return err
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)

			assertGroupGroupsEqual(t, dataStore.GroupGroups(), tt.wantGroupGroups)

			var groupAncestors []groupAncestor
			assert.NoError(t, dataStore.GroupAncestors().Select("group_ancestor_id, group_child_id, is_self").
				Order("group_ancestor_id, group_child_id").Scan(&groupAncestors).Error())

			sort.Slice(tt.wantGroupAncestors, func(i, j int) bool {
				return tt.wantGroupAncestors[i].AncestorGroupID < tt.wantGroupAncestors[j].AncestorGroupID ||
					(tt.wantGroupAncestors[i].AncestorGroupID == tt.wantGroupAncestors[j].AncestorGroupID &&
						tt.wantGroupAncestors[i].ChildGroupID < tt.wantGroupAncestors[j].ChildGroupID)
			})

			assert.Equal(t, tt.wantGroupAncestors, groupAncestors)

			var count int64
			assert.NoError(t, dataStore.Table("groups_propagate").
				Where("ancestors_computation_state != 'done'").Count(&count).Error())
			if tt.shouldRunListeners {
				assert.Zero(t, count, "Listeners should be executed")
			} else {
				assert.NotZero(t, count, "Listeners should not be executed")
			}
		})
	}
}

func patchGroupGroups(old []groupGroup, cycleWithType database.GroupGroupType, diff map[string]*groupGroup,
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
	if cycleWithType != "" {
		result = append(result,
			groupGroup{ParentGroupID: 20, ChildGroupID: 20, Type: string(cycleWithType)},
			groupGroup{ParentGroupID: 20, ChildGroupID: 30, Type: string(cycleWithType)},
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
		result[i] = invalid
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
	assert.NoError(t, groupGroupStore.Select("group_parent_id, group_child_id, user_inviting_id, child_order, type, status_date").
		Order("group_parent_id, group_child_id").Scan(&groupsGroups).Error())

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
		assert.Equal(t, row.Type, groupsGroups[index].Type, "wrong type for row %#v", groupsGroups[index])
		assert.Equal(t, row.InvitingUserID, groupsGroups[index].InvitingUserID, "wrong user_inviting_id for row %#v", groupsGroups[index])
		if row.ChildOrder == 0 {
			assert.Zero(t, groupsGroups[index].ChildOrder)
		} else {
			assert.False(t, usedChildOrders[groupsGroups[index].ChildOrder])
			usedChildOrders[groupsGroups[index].ChildOrder] = true
		}
		if row.StatusDate == nil {
			assert.Nil(t, groupsGroups[index].StatusDate)
		} else {
			assert.NotNil(t, groupsGroups[index].StatusDate, "StatusDate should not be nil in row %#v", groupsGroups[index])
			if groupsGroups[index].StatusDate != nil {
				assert.True(t, (*time.Time)(groupsGroups[index].StatusDate).Sub(time.Now().UTC())/time.Second < 5)
				assert.True(t, time.Now().UTC().Sub(time.Time(*groupsGroups[index].StatusDate))/time.Second > -5)
			}
		}
	}
}
