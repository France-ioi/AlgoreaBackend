// +build !unit

package database_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestUserStore_DeleteTemporaryWithTraps(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format("2006-01-02T15:04:05"))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.NoError(t, store.Users().DeleteTemporaryWithTraps())

	assertTableColumn(t, db, "users", "group_id", []int64{5001, 5002})
	assertTableColumn(t, db, "groups", "id", []int64{1, 5001, 5002, 7000})
	assertTableColumn(t, db, "groups_propagate", "id", []int64{1, 5001, 5002, 7000})
	assertTableColumn(t, db, "groups_ancestors", "ancestor_group_id", []int64{1, 5001, 5002, 7000})
	assertTableColumn(t, db, "groups_ancestors", "child_group_id", []int64{1, 5001, 5002, 7000})
	assertTableColumn(t, db, "groups_groups", "parent_group_id", []int64{1, 5001, 5002})
	assertTableColumn(t, db, "groups_groups", "child_group_id", []int64{5001, 5002, 7000})
	assertTableColumn(t, db, "group_pending_requests", "group_id", []int64{1, 5001, 5002})
	assertTableColumn(t, db, "group_pending_requests", "member_id", []int64{5001, 5002, 7000})
	assertTableColumn(t, db, "group_membership_changes", "group_id", []int64{1, 5001, 5002})
	assertTableColumn(t, db, "group_membership_changes", "member_id", []int64{5001, 5002, 7000})
	for _, table := range []string{"permissions_granted", "permissions_generated"} {
		assertTableColumn(t, db, table, "group_id", []int64{5001, 5002})
	}
	for _, table := range []string{"attempts", "results"} {
		assertTableColumn(t, db, table, "participant_id", []int64{5001, 5002})
	}
	assertTableColumn(t, db, "sessions", "user_id", []int64{5001})
	assertTableColumn(t, db, "answers", "author_id", []int64{5001, 5002})
	for _, table := range []string{"users_threads", "filters", "refresh_tokens"} {
		assertTableColumn(t, db, table, "user_id", []int64{5001, 5002})
	}

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func TestUserStore_DeleteWithTraps(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format("2006-01-02T15:04:05"))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.NoError(t, store.Users().DeleteWithTraps(
		&database.User{GroupID: 5001}))

	assertTableColumn(t, db, "users", "group_id", []int64{5000, 5002})
	assertTableColumn(t, db, "groups", "id", []int64{1, 5000, 5002, 7000})
	assertTableColumn(t, db, "groups_propagate", "id", []int64{1, 5000, 5002, 7000})
	assertTableColumn(t, db, "groups_ancestors", "ancestor_group_id", []int64{1, 5000, 5002, 7000})
	assertTableColumn(t, db, "groups_ancestors", "child_group_id", []int64{1, 5000, 5002, 7000})
	assertTableColumn(t, db, "groups_groups", "parent_group_id", []int64{1, 5000, 5002})
	assertTableColumn(t, db, "groups_groups", "child_group_id", []int64{5000, 5002, 7000})
	assertTableColumn(t, db, "group_pending_requests", "group_id", []int64{1, 5000, 5002})
	assertTableColumn(t, db, "group_pending_requests", "member_id", []int64{5000, 5002, 7000})
	assertTableColumn(t, db, "group_membership_changes", "group_id", []int64{1, 5000, 5002})
	assertTableColumn(t, db, "group_membership_changes", "member_id", []int64{5000, 5002, 7000})
	for _, table := range []string{"permissions_generated", "permissions_granted"} {
		assertTableColumn(t, db, table, "group_id", []int64{5000, 5002})
	}
	for _, table := range []string{"attempts", "results"} {
		assertTableColumn(t, db, table, "participant_id", []int64{5000, 5002})
	}
	assertTableColumn(t, db, "sessions", "user_id", []int64{5000})
	assertTableColumn(t, db, "answers", "author_id", []int64{5000, 5002})
	for _, table := range []string{"users_threads", "filters", "refresh_tokens"} {
		assertTableColumn(t, db, table, "user_id", []int64{5000, 5002})
	}

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func TestUserStore_DeleteWithTrapsByScope(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format("2006-01-02T15:04:05"))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.NoError(t, store.Users().DeleteWithTrapsByScope(func(store *database.DataStore) *database.DB {
		return store.Users().Where("group_id % 2 = 0")
	}))

	assertTableColumn(t, db, "users", "group_id", []int64{5001})
	assertTableColumn(t, db, "groups", "id", []int64{1, 5001, 7000})
	assertTableColumn(t, db, "groups_propagate", "id", []int64{1, 5001, 7000})
	assertTableColumn(t, db, "groups_ancestors", "ancestor_group_id", []int64{1, 5001, 7000})
	assertTableColumn(t, db, "groups_ancestors", "child_group_id", []int64{1, 5001, 7000})
	assertTableColumn(t, db, "groups_groups", "parent_group_id", []int64{1, 5001})
	assertTableColumn(t, db, "groups_groups", "child_group_id", []int64{5001, 7000})
	assertTableColumn(t, db, "group_pending_requests", "group_id", []int64{1, 5001})
	assertTableColumn(t, db, "group_pending_requests", "member_id", []int64{5001, 7000})
	assertTableColumn(t, db, "group_membership_changes", "group_id", []int64{1, 5001})
	assertTableColumn(t, db, "group_membership_changes", "member_id", []int64{5001, 7000})
	for _, table := range []string{"permissions_generated", "permissions_granted"} {
		assertTableColumn(t, db, table, "group_id", []int64{5001})
	}
	for _, table := range []string{"attempts", "results"} {
		assertTableColumn(t, db, table, "participant_id", []int64{5001})
	}
	assertTableColumn(t, db, "sessions", "user_id", []int64{5001})
	assertTableColumn(t, db, "answers", "author_id", []int64{5001})
	for _, table := range []string{"users_threads", "filters", "refresh_tokens"} {
		assertTableColumn(t, db, table, "user_id", []int64{5001})
	}

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func setupDBForDeleteWithTrapsTests(t *testing.T, currentTime time.Time) *database.DB {
	db := testhelpers.SetupDBWithFixtureString(`
			groups_propagate: [{id: 5000}, {id: 5001}, {id: 5002}]`, `
			groups: [{id: 1}, {id: 5000}, {id: 5001}, {id: 5002}, {id: 7000}]
			attempts:
				- {id: 1, participant_id: 5000}
				- {id: 1, participant_id: 5001}
				- {id: 1, participant_id: 5002}
			results:
				- {attempt_id: 1, participant_id: 5000, item_id: 1}
				- {attempt_id: 1, participant_id: 5001, item_id: 1}
				- {attempt_id: 1, participant_id: 5002, item_id: 1}
			users:
				- {temp_user: 1, login: 500, group_id: 5000} # should be deleted
				- {login: 501, temp_user: 1, group_id: 5001}
				- {login: 502, group_id: 5002}
			items: [{id: 1, default_language_tag: fr}]`, `
			sessions:
				- {user_id: 5000, expires_at: "`+currentTime.Format("2006-01-02 15:04:05")+`"}
				- {user_id: 5001, expires_at: "`+currentTime.Add(1*time.Second).Format("2006-01-02 15:04:05")+`"}
				- {user_id: 5001, expires_at: "`+currentTime.Add(-10*time.Second).Format("2006-01-02 15:04:05")+`"}
			users_threads: [{user_id: 5000, thread_id: 1}, {user_id: 5001, thread_id: 1}, {user_id: 5002, thread_id: 1}]
			answers:
				- {author_id: 5000, attempt_id: 1, participant_id: 5000, item_id: 1, created_at: 2019-05-30 11:00:00}
				- {author_id: 5001, attempt_id: 1, participant_id: 5001, item_id: 1, created_at: 2019-05-30 11:00:00}
				- {author_id: 5002, attempt_id: 1, participant_id: 5002, item_id: 1, created_at: 2019-05-30 11:00:00}
			filters: [{user_id: 5000}, {user_id: 5001}, {user_id: 5002}]
			refresh_tokens: [{user_id: 5000, refresh_token: token}, {user_id: 5001, refresh_token: token2},
			                 {user_id: 5002, refresh_token: token3}]
			permissions_generated:
				- {group_id: 5000, item_id: 1}
				- {group_id: 5001, item_id: 1}
				- {group_id: 5002, item_id: 1}
			permissions_granted:
				- {group_id: 5000, item_id: 1, source_group_id: 1}
				- {group_id: 5001, item_id: 1, source_group_id: 1}
				- {group_id: 5002, item_id: 1, source_group_id: 1}
			groups_groups: [{parent_group_id: 5000, child_group_id: 7000}, {parent_group_id: 5001, child_group_id: 7000},
			                {parent_group_id: 5002, child_group_id: 7000},
			                {parent_group_id: 1, child_group_id: 5000}, {parent_group_id: 1, child_group_id: 5001},
			                {parent_group_id: 1, child_group_id: 5002}]
			group_pending_requests: [{group_id: 5000, member_id: 7000}, {group_id: 5001, member_id: 7000},
			                         {group_id: 5002, member_id: 7000},
			                         {group_id: 1, member_id: 5000}, {group_id: 1, member_id: 5001},
			                         {group_id: 1, member_id: 5002}]
			group_membership_changes: [{group_id: 5000, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 5001, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 5002, member_id: 7000, at: "2019-05-30 11:00:00"},
			                           {group_id: 1, member_id: 5000}, {group_id: 1, member_id: 5001},
			                           {group_id: 1, member_id: 5002}]
			groups_ancestors: [{ancestor_group_id: 1, child_group_id: 1},
			                   {ancestor_group_id: 5000, child_group_id: 7000}, {ancestor_group_id: 5001, child_group_id: 7000},
			                   {ancestor_group_id: 5002, child_group_id: 7000},
			                   {ancestor_group_id: 1, child_group_id: 5000}, {ancestor_group_id: 1, child_group_id: 5001},
			                   {ancestor_group_id: 1, child_group_id: 5002},
			                   {ancestor_group_id: 1, child_group_id: 7000}]`)
	store := database.NewDataStore(db)
	assert.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
		trStore.GroupGroups().CreateNewAncestors()
		trStore.PermissionsGranted().ComputeAllAccess()
		return nil
	}))
	return db
}

func assertTableColumn(t *testing.T, db *database.DB, table, column string, expectedValues interface{}) {
	reflValues := reflect.New(reflect.TypeOf(expectedValues))
	assert.NoError(t, db.Table(table).Order(column).Pluck("DISTINCT "+column, reflValues.Interface()).Error())
	assert.EqualValues(t, expectedValues, reflValues.Elem().Interface(), "wrong %s in %s", column, table)
}
