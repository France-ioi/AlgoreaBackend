//go:build !unit

package database_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestUserStore_DeleteTemporaryWithTraps(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	for _, test := range []struct {
		name                                      string
		delay                                     time.Duration
		expectDeletedUsers, expectDeletedSessions *golang.Set[int64]
	}{
		{
			name:                  "no delay",
			delay:                 0,
			expectDeletedUsers:    golang.NewSet[int64](5000, 5003),
			expectDeletedSessions: golang.NewSet[int64](1),
		},
		{
			name:                  "delay 1 second",
			delay:                 1 * time.Second,
			expectDeletedUsers:    golang.NewSet[int64](5003),
			expectDeletedSessions: golang.NewSet[int64](),
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db := setupDBForDeleteWithTrapsTests(t, currentTime)
			defer func() { _ = db.Close() }()

			store := database.NewDataStore(db)
			require.NoError(t, store.Users().DeleteTemporaryWithTraps(test.delay))

			assertUserRelatedTablesAfterDeletingWithTraps(t, db, test.expectDeletedUsers, test.expectDeletedSessions)
		})
	}
}

func TestUserStore_DeleteWithTraps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	require.NoError(t, store.Users().DeleteWithTraps(
		&database.User{GroupID: 5001}, false))

	assertUserRelatedTablesAfterDeletingWithTraps(t, db, golang.NewSet[int64](5001), golang.NewSet[int64](2))
}

func TestUserStore_DeleteWithTrapsByScope(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format(time.DateTime))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	require.NoError(t, store.Users().DeleteWithTrapsByScope(func(store *database.DataStore) *database.DB {
		return store.Users().Where("group_id % 2 = 0")
	}, false))

	assertUserRelatedTablesAfterDeletingWithTraps(t, db, golang.NewSet[int64](5000, 5002), golang.NewSet[int64](1, 3))
}

func setupDBForDeleteWithTrapsTests(t *testing.T, currentTime time.Time) *database.DB {
	t.Helper()

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
			groups_propagate: [{id: 5000}, {id: 5001}, {id: 5002}, {id: 5003}]`, `
			groups: [{id: 1}, {id: 5000}, {id: 5001}, {id: 5002}, {id: 5003}, {id: 7000}]
			attempts:
				- {id: 1, participant_id: 5000}
				- {id: 1, participant_id: 5001}
				- {id: 1, participant_id: 5002}
				- {id: 1, participant_id: 5003}
			results:
				- {attempt_id: 1, participant_id: 5000, item_id: 1}
				- {attempt_id: 1, participant_id: 5001, item_id: 1}
				- {attempt_id: 1, participant_id: 5002, item_id: 1}
				- {attempt_id: 1, participant_id: 5003, item_id: 1}
			users:
				- {temp_user: 1, login: 500, group_id: 5000} # should be deleted, expired token
				- {login: 501, temp_user: 1, group_id: 5001} # should NOT be deleted, one access token is still valid
				- {login: 502, group_id: 5002} # should NOT be deleted, not a temporary user
				- {login: 503, temp_user: 1, group_id: 5003} # should be deleted, no session
			items: [{id: 1, default_language_tag: fr}]`, `
			sessions:
				- {session_id: 1, user_id: 5000, refresh_token: token}
				- {session_id: 2, user_id: 5001, refresh_token: token2}
				- {session_id: 3, user_id: 5002, refresh_token: token3}
			answers:
				- {author_id: 5000, attempt_id: 1, participant_id: 5000, item_id: 1, created_at: 2019-05-30 11:00:00}
				- {author_id: 5001, attempt_id: 1, participant_id: 5001, item_id: 1, created_at: 2019-05-30 11:00:00}
				- {author_id: 5002, attempt_id: 1, participant_id: 5002, item_id: 1, created_at: 2019-05-30 11:00:00}
				- {author_id: 5003, attempt_id: 1, participant_id: 5003, item_id: 1, created_at: 2019-05-30 11:00:00}
			filters: [{user_id: 5000}, {user_id: 5001}, {user_id: 5002}, {user_id: 5003}]
			access_tokens:
				- {session_id: 1, token: "token1", expires_at: "`+currentTime.Format(time.DateTime)+`"}
				- {session_id: 2, token: "token2a", expires_at: "`+currentTime.Add(1*time.Second).Format(time.DateTime)+`"}
				- {session_id: 2, token: "token2b", expires_at: "`+currentTime.Add(-10*time.Second).Format(time.DateTime)+`"}
				- {session_id: 3, token: "token3", expires_at: "`+currentTime.Add(-10*time.Second).Format(time.DateTime)+`"}
			permissions_generated:
				- {group_id: 5000, item_id: 1}
				- {group_id: 5001, item_id: 1}
				- {group_id: 5002, item_id: 1}
				- {group_id: 5003, item_id: 1}
			permissions_granted:
				- {group_id: 5000, item_id: 1, source_group_id: 1}
				- {group_id: 5001, item_id: 1, source_group_id: 1}
				- {group_id: 5002, item_id: 1, source_group_id: 1}
				- {group_id: 5003, item_id: 1, source_group_id: 1}
			groups_groups: [{parent_group_id: 5000, child_group_id: 7000}, {parent_group_id: 5001, child_group_id: 7000},
			                {parent_group_id: 5002, child_group_id: 7000}, {parent_group_id: 5003, child_group_id: 7000},
			                {parent_group_id: 1, child_group_id: 5000}, {parent_group_id: 1, child_group_id: 5001},
			                {parent_group_id: 1, child_group_id: 5002}, {parent_group_id: 1, child_group_id: 5003}]
			group_pending_requests: [{group_id: 5000, member_id: 7000}, {group_id: 5001, member_id: 7000},
			                         {group_id: 5002, member_id: 7000}, {group_id: 5003, member_id: 7000},
			                         {group_id: 1, member_id: 5000}, {group_id: 1, member_id: 5001},
			                         {group_id: 1, member_id: 5002}, {group_id: 1, member_id: 5003}]
			group_membership_changes: [{group_id: 5000, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 5001, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 5002, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 5003, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 1, member_id: 5000}, {group_id: 1, member_id: 5001},
			                           {group_id: 1, member_id: 5002}, {group_id: 1, member_id: 5003}]
			groups_ancestors: [{ancestor_group_id: 5000, child_group_id: 7000}, {ancestor_group_id: 5001, child_group_id: 7000},
			                   {ancestor_group_id: 5002, child_group_id: 7000}, {ancestor_group_id: 5003, child_group_id: 7000},
			                   {ancestor_group_id: 1, child_group_id: 5000}, {ancestor_group_id: 1, child_group_id: 5001},
			                   {ancestor_group_id: 1, child_group_id: 5002}, {ancestor_group_id: 1, child_group_id: 5003},
			                   {ancestor_group_id: 1, child_group_id: 7000}]`)
	store := database.NewDataStore(db)
	assert.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
		if err := trStore.GroupGroups().CreateNewAncestors(); err != nil {
			return err
		}
		trStore.SchedulePermissionsPropagation()
		return nil
	}))
	return db
}

func TestUserStore_DeleteWithTrapsByScope_RecomputesAccessWhenPermissionIsRemovedViaSourceGroupID(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db := testhelpers.SetupDBWithFixtureString(testhelpers.CreateTestContext(), `
		groups: [{id: 1}, {id: 5000}, {id: 5001}]
		users: [{group_id: 5000}]
		items: [{id: 10, default_language_tag: fr}]
		permissions_granted:
			- {group_id: 1, item_id: 10, source_group_id: 5000, can_view: content}
			- {group_id: 1, item_id: 10, source_group_id: 5001, can_view: info}
		permissions_generated: [{group_id: 1, item_id: 10, can_view_generated: content}]`)
	defer func() { _ = db.Close() }()

	dataStore := database.NewDataStore(db)
	require.NoError(t, dataStore.Users().DeleteWithTrapsByScope(func(store *database.DataStore) *database.DB {
		return store.Users().Where("group_id = 5000")
	}, false))

	var newPermission string
	require.NoError(t, dataStore.Permissions().Where("group_id = 1 AND item_id = 10").
		PluckFirst("can_view_generated", &newPermission).Error())
	assert.Equal(t, "info", newPermission)
}

func assertTableColumn(t *testing.T, db *database.DB, table, column string, expectedValues interface{}) {
	t.Helper()

	reflValues := reflect.New(reflect.TypeOf(expectedValues))
	require.NoError(t, db.Table(table).Order(column).Pluck("DISTINCT "+column, reflValues.Interface()).Error())
	assert.EqualValues(t, expectedValues, reflValues.Elem().Interface(), "wrong %s in %s", column, table)
}

func assertUserRelatedTablesAfterDeletingWithTraps(
	t *testing.T, db *database.DB, expectDeletedUsers, expectDeletedSessions *golang.Set[int64],
) {
	t.Helper()

	allUsers := filterIDs([]int64{5000, 5001, 5002, 5003}, expectDeletedUsers)
	allGroups := filterIDs([]int64{1, 5000, 5001, 5002, 5003, 7000}, expectDeletedUsers)
	allParentGroups := filterIDs([]int64{1, 5000, 5001, 5002, 5003}, expectDeletedUsers)
	allChildGroups := filterIDs([]int64{5000, 5001, 5002, 5003, 7000}, expectDeletedUsers)
	allPendingRequestGroups := filterIDs([]int64{1, 5000, 5001, 5002, 5003}, expectDeletedUsers)
	allPendingRequestMembers := filterIDs([]int64{5000, 5001, 5002, 5003, 7000}, expectDeletedUsers)
	otherTablesGroups := filterIDs([]int64{5000, 5001, 5002, 5003}, expectDeletedUsers)
	userIDsInSessions := filterIDs([]int64{5000, 5001, 5002}, expectDeletedUsers)
	sessions := filterIDs([]int64{1, 2, 3}, expectDeletedSessions)

	assertTableColumn(t, db, "users", "group_id", allUsers)
	assertTableColumn(t, db, "groups", "id", allGroups)
	assertTableColumn(t, db, "groups_ancestors", "ancestor_group_id", allGroups)
	assertTableColumn(t, db, "groups_ancestors", "child_group_id", allGroups)
	assertTableColumn(t, db, "groups_groups", "parent_group_id", allParentGroups)
	assertTableColumn(t, db, "groups_groups", "child_group_id", allChildGroups)
	assertTableColumn(t, db, "group_pending_requests", "group_id", allPendingRequestGroups)
	assertTableColumn(t, db, "group_pending_requests", "member_id", allPendingRequestMembers)
	assertTableColumn(t, db, "group_membership_changes", "group_id", allPendingRequestGroups)
	assertTableColumn(t, db, "group_membership_changes", "member_id", allPendingRequestMembers)
	for _, table := range []string{"permissions_granted", "permissions_generated"} {
		assertTableColumn(t, db, table, "group_id", otherTablesGroups)
	}
	for _, table := range []string{"attempts", "results"} {
		assertTableColumn(t, db, table, "participant_id", otherTablesGroups)
	}
	assertTableColumn(t, db, "sessions", "user_id", userIDsInSessions)
	assertTableColumn(t, db, "access_tokens", "session_id", sessions)
	assertTableColumn(t, db, "answers", "author_id", otherTablesGroups)
	assertTableColumn(t, db, "filters", "user_id", otherTablesGroups)

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})

	store := database.NewDataStore(db)
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	require.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func filterIDs(ids []int64, idsToExclude *golang.Set[int64]) (filtered []int64) {
	filtered = make([]int64, 0, len(ids))
	for _, id := range ids {
		if !idsToExclude.Contains(id) {
			filtered = append(filtered, id)
		}
	}
	return
}
