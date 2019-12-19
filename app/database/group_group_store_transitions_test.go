package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGroupGroupStore_transition_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_, _ = NewDataStore(db).GroupGroups().Transition(
			AdminCreatesInvitation, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, 12,
		)
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupGroupStore_transition_UsesNamedLock(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	dbMock.ExpectBegin()
	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(0)))
	dbMock.ExpectRollback()

	_ = NewDataStore(db).InTransaction(func(dataStore *DataStore) (err error) {
		_, err = dataStore.GroupGroups().Transition(
			AdminCreatesInvitation, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil, 12,
		)
		return
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
