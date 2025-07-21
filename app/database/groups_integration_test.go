//go:build !unit

package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDataStore_GetGroupJoiningByCodeInfoByCode(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	tests := []struct {
		name      string
		code      string
		withLock  bool
		wantFound bool
		want      database.GroupJoiningByCodeInfo
	}{
		{name: "wrong code", code: "bcd"},
		{name: "wrong code (the check is case-sensitive)", code: "UVWX"},
		{name: "wrong code (wildcards do not work)", code: "%"},
		{name: "wrong code (group is a user)", code: "xyza"},
		{
			name:      "group is not a team",
			code:      "abcd",
			withLock:  true,
			wantFound: true,
			want: database.GroupJoiningByCodeInfo{
				GroupID:             1,
				CodeExpiresAtIsNull: true,
				CodeLifetimeIsNull:  true,
				Type:                "Class",
			},
		},
		{
			name:      "group is not public",
			code:      "efgh",
			wantFound: true,
			want: database.GroupJoiningByCodeInfo{
				GroupID:             2,
				CodeExpiresAtIsNull: true,
				CodeLifetimeIsNull:  true,
				Type:                "Team",
			},
		},
		{name: "expired code", code: "ijkl"},
		{
			name:      "ok",
			code:      "mnop",
			withLock:  true,
			wantFound: true,
			want: database.GroupJoiningByCodeInfo{
				GroupID:             4,
				CodeExpiresAtIsNull: false,
				CodeLifetimeIsNull:  true,
				FrozenMembership:    false,
				Type:                "Team",
			},
		},
		{
			name:      "ok (expires at is null)",
			code:      "qrst",
			wantFound: true,
			want: database.GroupJoiningByCodeInfo{
				GroupID:             5,
				CodeExpiresAtIsNull: true,
				CodeLifetimeIsNull:  false,
				FrozenMembership:    false,
				Type:                "Team",
			},
		},
		{
			name:      "ok (frozen membership)",
			code:      "uvwx",
			withLock:  true,
			wantFound: true,
			want: database.GroupJoiningByCodeInfo{
				GroupID:             6,
				CodeExpiresAtIsNull: true,
				CodeLifetimeIsNull:  true,
				FrozenMembership:    true,
				Type:                "Team",
			},
		},
	}

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups:
			- {id: 1, type: Class, code: abcd, is_public: 1}
			- {id: 2, type: Team, code: efgh}
			- {id: 3, type: Team, code: ijkl, is_public: 1, code_expires_at: 2019-05-30 11:00:00}
			- {id: 4, type: Team, code: mnop, is_public: 1, code_expires_at: 3019-05-30 11:00:00}
			- {id: 5, type: Team, code: qrst, is_public: 1, code_lifetime: 3600}
			- {id: 6, type: Team, code: uvwx, is_public: 1, frozen_membership: 1}
			- {id: 7, type: User, code: xyza}
		`)
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			store := database.NewDataStore(db)
			var got database.GroupJoiningByCodeInfo
			var found bool
			var err error
			if tt.withLock {
				require.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
					got, found, err = trStore.GetGroupJoiningByCodeInfoByCode(tt.code, tt.withLock)
					return err
				}))
			} else {
				got, found, err = store.GetGroupJoiningByCodeInfoByCode(tt.code, tt.withLock)
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.want, got)
		})
	}
}
