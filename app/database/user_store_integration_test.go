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
	testhelpers.MockDBTime(currentTime.Format(time.RFC3339))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.NoError(t, store.Users().DeleteTemporaryWithTraps())

	for _, table := range []string{"users", "history_users"} {
		assertTableColumn(t, db, table, "id", []int64{501, 502})
	}
	for _, table := range []string{"groups", "history_groups"} {
		assertTableColumn(t, db, table, "id", []int64{1, 5001, 5002, 6001, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_propagate", "id", []int64{1, 5001, 5002, 6001, 6002, 7000})
	for _, table := range []string{"groups_ancestors", "history_groups_ancestors"} {
		assertTableColumn(t, db, table, "ancestor_group_id", []int64{1, 5001, 5002, 6001, 6002, 7000})
		assertTableColumn(t, db, table, "child_group_id", []int64{1, 5001, 5002, 6001, 6002, 7000})
	}
	for _, table := range []string{"groups_groups", "history_groups_groups"} {
		assertTableColumn(t, db, table, "parent_group_id", []int64{1, 5001, 5002, 6001, 6002})
		assertTableColumn(t, db, table, "child_group_id", []int64{5001, 5002, 6001, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_items_propagate", "id", []int64{2, 3, 5, 6})
	for _, table := range []string{
		"groups_items", "history_groups_items", "groups_attempts", "history_groups_attempts",
		"groups_login_prefixes", "history_groups_login_prefixes",
	} {
		assertTableColumn(t, db, table, "group_id", []int64{5001, 5002, 6001, 6002})
	}
	assertTableColumn(t, db, "sessions", "user_id", []int64{501})
	for _, table := range []string{
		"users_threads", "history_users_threads", "users_answers", "users_items", "history_users_items",
		"filters", "history_filters", "refresh_tokens",
	} {
		assertTableColumn(t, db, table, "user_id", []int64{501, 502})
	}

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func TestUserStore_DeleteWithTraps(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(time.Second)
	testhelpers.MockDBTime(currentTime.Format(time.RFC3339))
	defer testhelpers.RestoreDBTime()

	db := setupDBForDeleteWithTrapsTests(t, currentTime)
	defer func() { _ = db.Close() }()

	store := database.NewDataStore(db)
	assert.NoError(t, store.Users().DeleteWithTraps(
		&database.User{ID: 501, SelfGroupID: ptrInt64(5001), OwnedGroupID: ptrInt64(6001)}))

	for _, table := range []string{"users", "history_users"} {
		assertTableColumn(t, db, table, "id", []int64{500, 502})
	}
	for _, table := range []string{"groups", "history_groups"} {
		assertTableColumn(t, db, table, "id", []int64{1, 5000, 5002, 6000, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_propagate", "id", []int64{1, 5000, 5002, 6000, 6002, 7000})
	for _, table := range []string{"groups_ancestors", "history_groups_ancestors"} {
		assertTableColumn(t, db, table, "ancestor_group_id", []int64{1, 5000, 5002, 6000, 6002, 7000})
		assertTableColumn(t, db, table, "child_group_id", []int64{1, 5000, 5002, 6000, 6002, 7000})
	}
	for _, table := range []string{"groups_groups", "history_groups_groups"} {
		assertTableColumn(t, db, table, "parent_group_id", []int64{1, 5000, 5002, 6000, 6002})
		assertTableColumn(t, db, table, "child_group_id", []int64{5000, 5002, 6000, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_items_propagate", "id", []int64{1, 3, 4, 6})
	for _, table := range []string{
		"groups_items", "history_groups_items", "groups_attempts", "history_groups_attempts",
		"groups_login_prefixes", "history_groups_login_prefixes",
	} {
		assertTableColumn(t, db, table, "group_id", []int64{5000, 5002, 6000, 6002})
	}
	assertTableColumn(t, db, "sessions", "user_id", []int64{500})
	for _, table := range []string{
		"users_threads", "history_users_threads", "users_answers", "users_items", "history_users_items",
		"filters", "history_filters", "refresh_tokens",
	} {
		assertTableColumn(t, db, table, "user_id", []int64{500, 502})
	}

	assertTableColumn(t, db, "groups_propagate", "ancestors_computation_state", []string{"done"})
	found, err := store.GroupAncestors().Where("ancestor_group_id = 1 AND child_group_id = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func setupDBForDeleteWithTrapsTests(t *testing.T, currentTime time.Time) *database.DB {
	db := testhelpers.SetupDBWithFixtureString(`
			groups_propagate: [{id: 5000}, {id: 5001}, {id: 5002}, {id: 6000}, {id: 6001}, {id: 6002}]
			history_users: [{id: 500, version: 1}, {id: 501, version: 1}, {id: 502, version: 1}]
			`, `
			users:
				- {id: 500, temp_user: 1, login: 500, self_group_id: 5000, owned_group_id: 6000} # should be deleted
				- {id: 501, login: 501, temp_user: 1, self_group_id: 5001, owned_group_id: 6001}
				- {id: 502, login: 502, self_group_id: 5002, owned_group_id: 6002}
			sessions:
				- {user_id: 500, expiration_date: "`+currentTime.Format("2006-01-02 15:04:05")+`"}
				- {user_id: 501, expiration_date: "`+currentTime.Add(1*time.Second).Format("2006-01-02 15:04:05")+`"}
				- {user_id: 501, expiration_date: "`+currentTime.Add(-10*time.Second).Format("2006-01-02 15:04:05")+`"}
			users_threads: [{user_id: 500, thread_id: 1}, {user_id: 501, thread_id: 1}, {user_id: 502, thread_id: 1}]
			history_users_threads: [{user_id: 500, id: 1, thread_id: 1, version: 1}, {user_id: 501, id: 2, thread_id: 2, version: 1},
			                        {user_id: 502, id: 3, thread_id: 3, version: 1}]
			users_answers:
				- {user_id: 500, item_id: 1, submission_date: 2019-05-30 11:00:00}
				- {user_id: 501, item_id: 1, submission_date: 2019-05-30 11:00:00}
				- {user_id: 502, item_id: 1, submission_date: 2019-05-30 11:00:00}
			users_items: [{user_id: 500, item_id: 1}, {user_id: 501, item_id: 1}, {user_id: 502, item_id: 1}]
			history_users_items:
				- {user_id: 500, item_id: 1, id: 1, submissions_attempts: 0, tasks_tried: 0, children_validated: 0,
				   validated: 0, finished: 0, tasks_with_help: 0, hints_cached: 0, corrections_read: 0, precision: 0,
				   autonomy: 0, ranked: 0, version: 1}
				- {user_id: 501, item_id: 1, id: 2, submissions_attempts: 0, tasks_tried: 0, children_validated: 0,
				   validated: 0, finished: 0, tasks_with_help: 0, hints_cached: 0, corrections_read: 0,
				   precision: 0, autonomy: 0, ranked: 0, version: 1}
				- {user_id: 502, item_id: 1, id: 3, submissions_attempts: 0, tasks_tried: 0, children_validated: 0,
				   validated: 0, finished: 0, tasks_with_help: 0, hints_cached: 0, corrections_read: 0, precision: 0,
				   autonomy: 0, ranked: 0, version: 1}
			filters: [{user_id: 500}, {user_id: 501}, {user_id: 502}]
			history_filters: [{user_id: 500, id: 1, version: 1}, {user_id: 501, id: 2, version: 1}, {user_id: 502, id: 3, version: 1}]
			refresh_tokens: [{user_id: 500, refresh_token: token}, {user_id: 501, refresh_token: token2},
			                 {user_id: 502, refresh_token: token3}]
			groups_items_propagate: [{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}, {id: 6}]
			groups_items:
				- {id: 1, group_id: 5000, item_id: 1, creator_user_id: 1}
				- {id: 2, group_id: 5001, item_id: 1, creator_user_id: 1}
				- {id: 3, group_id: 5002, item_id: 1, creator_user_id: 1}
				- {id: 4, group_id: 6000, item_id: 1, creator_user_id: 1}
				- {id: 5, group_id: 6001, item_id: 1, creator_user_id: 1}
				- {id: 6, group_id: 6002, item_id: 1, creator_user_id: 1}
			history_groups_items:
				- {id: 1, group_id: 5000, item_id: 1, creator_user_id: 1, version: 1}
				- {id: 2, group_id: 5001, item_id: 1, creator_user_id: 1, version: 1}
				- {id: 3, group_id: 5002, item_id: 1, creator_user_id: 1, version: 1}
				- {id: 4, group_id: 6000, item_id: 1, creator_user_id: 1, version: 1}
				- {id: 5, group_id: 6001, item_id: 1, creator_user_id: 1, version: 1}
				- {id: 6, group_id: 6002, item_id: 1, creator_user_id: 1, version: 1}
			groups_attempts: [{group_id: 5000, item_id: 1, order: 0}, {group_id: 5001, item_id: 1, order: 0},
			                  {group_id: 5002, item_id: 1, order: 0}, {group_id: 6000, item_id: 1, order: 0},
			                  {group_id: 6001, item_id: 1, order: 0}, {group_id: 6002, item_id: 1, order: 0}]
			history_groups_attempts:
				- {group_id: 5000, item_id: 1, id: 1, order: 0, version: 1}
				- {group_id: 5001, item_id: 1, id: 2, order: 0, version: 1}
				- {group_id: 5002, item_id: 1, id: 3, order: 0, version: 1}
				- {group_id: 6000, item_id: 1, id: 4, order: 0, version: 1}
				- {group_id: 6001, item_id: 1, id: 5, order: 0, version: 1}
				- {group_id: 6002, item_id: 1, id: 6, order: 0, version: 1}
			groups_login_prefixes: [{group_id: 5000, prefix: 5000, id: 1}, {group_id: 5001, prefix: 5001, id: 2},
			                        {group_id: 5002, prefix: 5002, id: 3}, {group_id: 6000, prefix: 6000, id: 4},
			                        {group_id: 6001, prefix: 6001, id: 5}, {group_id: 6002, prefix: 6002, id: 6}]
			history_groups_login_prefixes:
				- {group_id: 5000, prefix: 5000, id: 1, version: 1}
				- {group_id: 5001, prefix: 5001, id: 2, version: 1}
				- {group_id: 5002, prefix: 5002, id: 3, version: 1}
				- {group_id: 6000, prefix: 5003, id: 4, version: 1}
				- {group_id: 6001, prefix: 5004, id: 5, version: 1}
				- {group_id: 6002, prefix: 5004, id: 6, version: 1}
			groups_groups: [{parent_group_id: 5000, child_group_id: 7000}, {parent_group_id: 5001, child_group_id: 7000},
			                {parent_group_id: 5002, child_group_id: 7000}, {parent_group_id: 6000, child_group_id: 7000},
			                {parent_group_id: 6001, child_group_id: 7000}, {parent_group_id: 6002, child_group_id: 7000},
			                {parent_group_id: 1, child_group_id: 5000}, {parent_group_id: 1, child_group_id: 5001},
			                {parent_group_id: 1, child_group_id: 5002}, {parent_group_id: 1, child_group_id: 6000},
			                {parent_group_id: 1, child_group_id: 6001}, {parent_group_id: 1, child_group_id: 6002}]
			history_groups_groups:
				- {parent_group_id: 5000, child_group_id: 7000, id: 1, child_order: 0, version: 1}
				- {parent_group_id: 5001, child_group_id: 7000, id: 2, child_order: 0, version: 1}
				- {parent_group_id: 5002, child_group_id: 7000, id: 3, child_order: 0, version: 1}
				- {parent_group_id: 6000, child_group_id: 7000, id: 4, child_order: 0, version: 1}
				- {parent_group_id: 6001, child_group_id: 7000, id: 5, child_order: 0, version: 1}
				- {parent_group_id: 6002, child_group_id: 7000, id: 6, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 5000, id: 7, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 5001, id: 8, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 5002, id: 9, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 6000, id: 10, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 6001, id: 11, child_order: 0, version: 1}
				- {parent_group_id: 1, child_group_id: 6002, id: 12, child_order: 0, version: 1}
			groups_ancestors: [{ancestor_group_id: 1, child_group_id: 1},
			                   {ancestor_group_id: 5000, child_group_id: 7000}, {ancestor_group_id: 5001, child_group_id: 7000},
			                   {ancestor_group_id: 5002, child_group_id: 7000}, {ancestor_group_id: 6000, child_group_id: 7000},
			                   {ancestor_group_id: 6001, child_group_id: 7000}, {ancestor_group_id: 6002, child_group_id: 7000},
			                   {ancestor_group_id: 1, child_group_id: 5000}, {ancestor_group_id: 1, child_group_id: 5001},
			                   {ancestor_group_id: 1, child_group_id: 5002}, {ancestor_group_id: 1, child_group_id: 6000},
			                   {ancestor_group_id: 1, child_group_id: 6001}, {ancestor_group_id: 1, child_group_id: 6002},
			                   {ancestor_group_id: 1, child_group_id: 7000}]
			history_groups_ancestors:
				- {ancestor_group_id: 5000, child_group_id: 7000, id: 1, version: 1}
				- {ancestor_group_id: 5001, child_group_id: 7000, id: 2, version: 1}
				- {ancestor_group_id: 5002, child_group_id: 7000, id: 3, version: 1}
				- {ancestor_group_id: 6000, child_group_id: 7000, id: 4, version: 1}
				- {ancestor_group_id: 6001, child_group_id: 7000, id: 5, version: 1}
				- {ancestor_group_id: 6002, child_group_id: 7000, id: 6, version: 1}
				- {ancestor_group_id: 1, child_group_id: 5000, id: 7, version: 1}
				- {ancestor_group_id: 1, child_group_id: 5001, id: 8, version: 1}
				- {ancestor_group_id: 1, child_group_id: 5002, id: 8, version: 1}
				- {ancestor_group_id: 1, child_group_id: 6000, id: 9, version: 1}
				- {ancestor_group_id: 1, child_group_id: 6001, id: 10, version: 1}
				- {ancestor_group_id: 1, child_group_id: 6002, id: 11, version: 1}
			groups: [{id: 1}, {id: 5000}, {id: 5001}, {id: 5002}, {id: 6000}, {id: 6001}, {id: 6002}, {id: 7000}]
			history_groups:
				- {id: 5000, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 5001, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 5002, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 6000, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 6001, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 6002, opened: 0, free_access: 1, send_emails: 0, version: 1}
				- {id: 7000, opened: 0, free_access: 1, send_emails: 0, version: 1}`)
	store := database.NewDataStore(db)
	store.GroupGroups().CreateNewAncestors()
	assert.NoError(t, store.InTransaction(func(trStore *database.DataStore) error {
		trStore.GroupItems().ComputeAllAccess()
		return nil
	}))
	return db
}

func assertTableColumn(t *testing.T, db *database.DB, table, column string, expectedValues interface{}) {
	reflValues := reflect.New(reflect.TypeOf(expectedValues))
	assert.NoError(t, db.Table(table).Order(column).Pluck("DISTINCT "+column, reflValues.Interface()).Error())
	assert.Equal(t, expectedValues, reflValues.Elem().Interface(), "wrong %s in %s", column, table)
}
