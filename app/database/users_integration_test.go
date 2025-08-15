//go:build !unit

package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDataStore_CheckIfTeamParticipationsConflictWithExistingUserMemberships(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	tests := []struct {
		name   string
		teamID int64
		userID int64
		want   bool
	}{
		{
			name:   "no conflicts (participations of non-team groups & stale memberships/attempts are ignored)",
			teamID: 4,
			userID: 3,
			want:   false,
		},
		{
			name:   "conflict (multiple attempts are disallowed)",
			teamID: 5,
			userID: 3,
			want:   true,
		},
		{
			name:   "conflict (multiple attempts are disallowed 2)",
			teamID: 6,
			userID: 3,
			want:   true,
		},
		{
			name:   "conflict",
			teamID: 7,
			userID: 3,
			want:   true,
		},
	}
	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups:
			- {id: 1, type: Class}
			- {id: 2, type: Team}
			- {id: 3, type: User}
			- {id: 4, type: Team}
			- {id: 5, type: Team}
			- {id: 6, type: Team}
			- {id: 7, type: Team}
		users:
			- {group_id: 3, login: john}
		groups_groups:
			- {parent_group_id: 1, child_group_id: 3}
			- {parent_group_id: 2, child_group_id: 3}
			- {parent_group_id: 4, child_group_id: 3, expires_at: 2019-05-30 11:00:00}
		items:
			- {id: 100, default_language_tag: en, allows_multiple_attempts: true}
			- {id: 101, default_language_tag: en, allows_multiple_attempts: true}
			- {id: 102, default_language_tag: en, allows_multiple_attempts: false}
			- {id: 103, default_language_tag: en, allows_multiple_attempts: false}
		attempts:
			- {id: 201, participant_id: 1, root_item_id: 100}
			- {id: 202, participant_id: 2, root_item_id: 100, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 203, participant_id: 2, root_item_id: 101}
			- {id: 204, participant_id: 2, root_item_id: 102, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 205, participant_id: 2, root_item_id: 103}
			- {id: 202, participant_id: 3, root_item_id: 100}
			- {id: 203, participant_id: 3, root_item_id: 101}

			- {id: 203, participant_id: 4, root_item_id: null}
			- {id: 204, participant_id: 4, root_item_id: 100}
			- {id: 205, participant_id: 4, root_item_id: 101, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 203, participant_id: 5, root_item_id: null}
			- {id: 204, participant_id: 5, root_item_id: 100}
			- {id: 205, participant_id: 5, root_item_id: 101, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 206, participant_id: 5, root_item_id: 102}
			- {id: 204, participant_id: 6, root_item_id: 100}
			- {id: 205, participant_id: 6, root_item_id: 101, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 206, participant_id: 6, root_item_id: 103, allows_submissions_until: 2019-05-30 11:00:00}
			- {id: 204, participant_id: 7, root_item_id: 100}
			- {id: 205, participant_id: 7, root_item_id: 101, allows_submissions_until: 3019-05-30 11:00:00}
		`)
	defer func() { _ = db.Close() }()
	for _, tt := range tests {
		tt := tt
		for _, withLock := range []bool{true, false} {
			withLock := withLock
			t.Run(tt.name+fmt.Sprintf(" withLock = %v", withLock), func(t *testing.T) {
				testoutput.SuppressIfPasses(t)

				store := database.NewDataStore(db)
				var got bool
				var err error
				if withLock {
					require.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
						got, err = trStore.CheckIfTeamParticipationsConflictWithExistingUserMemberships(tt.teamID, tt.userID, true)
						return err
					}))
				} else {
					got, err = store.CheckIfTeamParticipationsConflictWithExistingUserMemberships(tt.teamID, tt.userID, false)
				}
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	}
}
