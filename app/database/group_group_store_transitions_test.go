package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupGroupStore_Transition_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _, _ = NewDataStore(db).GroupGroups().Transition(
			AdminCreatesInvitation, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, 12,
		)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestMembershipAction_PendingType(t *testing.T) {
	assert.Equal(t, "invitation", InvitationCreated.PendingType())
	assert.Equal(t, "join_request", JoinRequestCreated.PendingType())
	assert.Panics(t, func() { JoinRequestAccepted.PendingType() })
}

func TestGroupApprovals_FromString(t *testing.T) {
	for _, test := range []struct {
		name                   string
		csv                    string
		expectedGroupApprovals GroupApprovals
	}{
		{
			name: "all are set",
			csv:  "personal_info_view,lock_membership,watch",
			expectedGroupApprovals: GroupApprovals{
				PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true,
			},
		},
		{name: "none are set", csv: "wrong"},
		{
			name:                   "personal_info_view",
			csv:                    "personal_info_view",
			expectedGroupApprovals: GroupApprovals{PersonalInfoViewApproval: true},
		},
		{
			name:                   "lock_membership",
			csv:                    "lock_membership",
			expectedGroupApprovals: GroupApprovals{LockMembershipApproval: true},
		},
		{
			name:                   "watch",
			csv:                    "watch",
			expectedGroupApprovals: GroupApprovals{WatchApproval: true},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var approvals GroupApprovals
			approvals.FromString(test.csv)
			assert.Equal(t, test.expectedGroupApprovals, approvals)
		})
	}
}

func TestGroupApprovals_ToArray(t *testing.T) {
	for _, test := range []struct {
		name           string
		groupApprovals GroupApprovals
		expectedArray  []string
	}{
		{
			name: "all are set",
			groupApprovals: GroupApprovals{
				PersonalInfoViewApproval: true, LockMembershipApproval: true, WatchApproval: true,
			},
			expectedArray: []string{"personal_info_view", "lock_membership", "watch"},
		},
		{name: "none are set", expectedArray: []string{}},
		{
			name:           "personal_info_view",
			groupApprovals: GroupApprovals{PersonalInfoViewApproval: true},
			expectedArray:  []string{"personal_info_view"},
		},
		{
			name:           "lock_membership",
			groupApprovals: GroupApprovals{LockMembershipApproval: true},
			expectedArray:  []string{"lock_membership"},
		},
		{
			name:           "watch",
			groupApprovals: GroupApprovals{WatchApproval: true},
			expectedArray:  []string{"watch"},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedArray, test.groupApprovals.ToArray())
		})
	}
}
