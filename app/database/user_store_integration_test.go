// +build !unit

package database_test

import (
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

	db := testhelpers.SetupDBWithFixtureString(`
		groups_propagate: [{ID: 5000}, {ID: 5001}, {ID: 5002}, {ID: 6000}, {ID: 6001}, {ID: 6002}]
		history_users: [{ID: 500}, {ID: 501}, {ID: 502}]
		`, `
		users:
			- {ID: 500, tempUser: 1, sLogin: 500, idGroupSelf: 5000, idGroupOwned: 6000} # should be deleted
			- {ID: 501, sLogin: 501, tempUser: 1, idGroupSelf: 5001, idGroupOwned: 6001}
			- {ID: 502, sLogin: 502, idGroupSelf: 5002, idGroupOwned: 6002}
		sessions:
			- {idUser: 500, sExpirationDate: "`+currentTime.Format(time.RFC3339)+`"}
			- {idUser: 501, sExpirationDate: "`+currentTime.Add(1*time.Second).Format(time.RFC3339)+`"}
			- {idUser: 501, sExpirationDate: "`+currentTime.Add(-10*time.Second).Format(time.RFC3339)+`"}
		users_threads: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		history_users_threads: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		users_answers: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		users_items: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		history_users_items: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		filters: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		history_filters: [{idUser: 500}, {idUser: 501}, {idUser: 502}] 
		refresh_tokens: [{idUser: 500}, {idUser: 501}, {idUser: 502}]
		groups_items_propagate: [{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}] 
		groups_items: [{ID: 1, idGroup: 5000}, {ID: 2, idGroup: 5001}, {ID: 3, idGroup: 5002}, {ID: 4, idGroup: 6000},
		               {ID: 5, idGroup: 6001}, {ID: 6, idGroup: 6002}] 
		history_groups_items: [{ID: 1, idGroup: 5000}, {ID: 2, idGroup: 5001}, {ID: 3, idGroup: 5002}, {ID: 4, idGroup: 6000},
		                       {ID: 5, idGroup: 6001}, {ID: 6, idGroup: 6002}] 
		groups_attempts: [{idGroup: 5000}, {idGroup: 5001}, {idGroup: 5002}, {idGroup: 6000}, {idGroup: 6001}, {idGroup: 6002}] 
		history_groups_attempts: [{idGroup: 5000}, {idGroup: 5001}, {idGroup: 5002}, {idGroup: 6000}, {idGroup: 6001}, {idGroup: 6002}] 
		groups_login_prefixes: [{idGroup: 5000, prefix: 5000}, {idGroup: 5001, prefix: 5001}, {idGroup: 5002, prefix: 5002},
		                        {idGroup: 6000, prefix: 6000}, {idGroup: 6001, prefix: 6001}, {idGroup: 6002, prefix: 6002}] 
		history_groups_login_prefixes: [{idGroup: 5000}, {idGroup: 5001}, {idGroup: 5002}, {idGroup: 6000}, {idGroup: 6001}, {idGroup: 6002}] 
		groups_groups: [{idGroupParent: 5000, idGroupChild: 7000}, {idGroupParent: 5001, idGroupChild: 7000},
		                {idGroupParent: 5002, idGroupChild: 7000}, {idGroupParent: 6000, idGroupChild: 7000},
		                {idGroupParent: 6001, idGroupChild: 7000}, {idGroupParent: 6002, idGroupChild: 7000},
		                {idGroupParent: 1, idGroupChild: 5000}, {idGroupParent: 1, idGroupChild: 5001},
		                {idGroupParent: 1, idGroupChild: 5002}, {idGroupParent: 1, idGroupChild: 6000},
		                {idGroupParent: 1, idGroupChild: 6001}, {idGroupParent: 1, idGroupChild: 6002}] 
		history_groups_groups: [{idGroupParent: 5000, idGroupChild: 7000}, {idGroupParent: 5001, idGroupChild: 7000},
		                        {idGroupParent: 5002, idGroupChild: 7000}, {idGroupParent: 6000, idGroupChild: 7000},
		                        {idGroupParent: 6001, idGroupChild: 7000}, {idGroupParent: 6002, idGroupChild: 7000},
		                        {idGroupParent: 1, idGroupChild: 5000}, {idGroupParent: 1, idGroupChild: 5001},
		                        {idGroupParent: 1, idGroupChild: 5002}, {idGroupParent: 1, idGroupChild: 6000},
		                        {idGroupParent: 1, idGroupChild: 6001}, {idGroupParent: 1, idGroupChild: 6002}] 
		groups_ancestors: [{idGroupAncestor: 5000, idGroupChild: 7000}, {idGroupAncestor: 5001, idGroupChild: 7000},
		                   {idGroupAncestor: 5002, idGroupChild: 7000}, {idGroupAncestor: 6000, idGroupChild: 7000},
		                   {idGroupAncestor: 6001, idGroupChild: 7000}, {idGroupAncestor: 6002, idGroupChild: 7000},
		                   {idGroupAncestor: 1, idGroupChild: 5000}, {idGroupAncestor: 1, idGroupChild: 5001},
		                   {idGroupAncestor: 1, idGroupChild: 5002}, {idGroupAncestor: 1, idGroupChild: 6000},
		                   {idGroupAncestor: 1, idGroupChild: 6001}, {idGroupAncestor: 1, idGroupChild: 6002}] 
		history_groups_ancestors: [{idGroupAncestor: 5000, idGroupChild: 7000}, {idGroupAncestor: 5001, idGroupChild: 7000},
		                           {idGroupAncestor: 5002, idGroupChild: 7000}, {idGroupAncestor: 6000, idGroupChild: 7000},
		                           {idGroupAncestor: 6001, idGroupChild: 7000}, {idGroupAncestor: 6002, idGroupChild: 7000},
		                           {idGroupAncestor: 1, idGroupChild: 5000}, {idGroupAncestor: 1, idGroupChild: 5001},
		                           {idGroupAncestor: 1, idGroupChild: 5002}, {idGroupAncestor: 1, idGroupChild: 6000},
		                           {idGroupAncestor: 1, idGroupChild: 6001}, {idGroupAncestor: 1, idGroupChild: 6002}]
		groups: [{ID: 5000}, {ID: 5001}, {ID: 5002}, {ID: 6000}, {ID: 6001}, {ID: 6002}]   
		history_groups: [{ID: 5000}, {ID: 5001}, {ID: 5002}, {ID: 6000}, {ID: 6001}, {ID: 6002}]`)
	defer func() { _ = db.Close() }()

	assert.NoError(t, database.NewDataStore(db).Users().DeleteTemporaryWithTraps())

	for _, table := range []string{"users", "history_users"} {
		assertTableColumn(t, db, table, "ID", []int64{501, 502})
	}
	for _, table := range []string{"groups", "history_groups"} {
		assertTableColumn(t, db, table, "ID", []int64{5001, 5002, 6001, 6002})
	}
	assertTableColumn(t, db, "groups_propagate", "ID", []int64{1, 5001, 5002, 6001, 6002, 7000})
	for _, table := range []string{"groups_ancestors", "history_groups_ancestors"} {
		assertTableColumn(t, db, table, "idGroupAncestor", []int64{1, 5001, 5002, 6001, 6002})
		assertTableColumn(t, db, table, "idGroupChild", []int64{5001, 5002, 6001, 6002, 7000})
	}
	for _, table := range []string{"groups_groups", "history_groups_groups"} {
		assertTableColumn(t, db, table, "idGroupParent", []int64{1, 5001, 5002, 6001, 6002})
		assertTableColumn(t, db, table, "idGroupChild", []int64{5001, 5002, 6001, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_items_propagate", "ID", []int64{2, 3, 5, 6})
	for _, table := range []string{
		"groups_items", "history_groups_items", "groups_attempts", "history_groups_attempts",
		"groups_login_prefixes", "history_groups_login_prefixes",
	} {
		assertTableColumn(t, db, table, "idGroup", []int64{5001, 5002, 6001, 6002})
	}
	assertTableColumn(t, db, "sessions", "idUser", []int64{501})
	for _, table := range []string{
		"users_threads", "history_users_threads", "users_answers", "users_items", "history_users_items",
		"filters", "history_filters", "refresh_tokens",
	} {
		assertTableColumn(t, db, table, "idUser", []int64{501, 502})
	}
}

func assertTableColumn(t *testing.T, db *database.DB, table, column string, expectedValues []int64) {
	var ids []int64
	assert.NoError(t, db.Table(table).Order(column).Pluck("DISTINCT "+column, &ids).Error())
	assert.Equal(t, expectedValues, ids, "wrong %s in %s", column, table)
}
