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
		assertTableColumn(t, db, table, "ID", []int64{501, 502})
	}
	for _, table := range []string{"groups", "history_groups"} {
		assertTableColumn(t, db, table, "ID", []int64{1, 5001, 5002, 6001, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_propagate", "ID", []int64{1, 5001, 5002, 6001, 6002, 7000})
	for _, table := range []string{"groups_ancestors", "history_groups_ancestors"} {
		assertTableColumn(t, db, table, "idGroupAncestor", []int64{1, 5001, 5002, 6001, 6002, 7000})
		assertTableColumn(t, db, table, "idGroupChild", []int64{1, 5001, 5002, 6001, 6002, 7000})
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

	assertTableColumn(t, db, "groups_propagate", "sAncestorsComputationState", []string{"done"})
	found, err := store.GroupAncestors().Where("idGroupAncestor = 1 AND idGroupChild = 7000").HasRows()
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
		assertTableColumn(t, db, table, "ID", []int64{500, 502})
	}
	for _, table := range []string{"groups", "history_groups"} {
		assertTableColumn(t, db, table, "ID", []int64{1, 5000, 5002, 6000, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_propagate", "ID", []int64{1, 5000, 5002, 6000, 6002, 7000})
	for _, table := range []string{"groups_ancestors", "history_groups_ancestors"} {
		assertTableColumn(t, db, table, "idGroupAncestor", []int64{1, 5000, 5002, 6000, 6002, 7000})
		assertTableColumn(t, db, table, "idGroupChild", []int64{1, 5000, 5002, 6000, 6002, 7000})
	}
	for _, table := range []string{"groups_groups", "history_groups_groups"} {
		assertTableColumn(t, db, table, "idGroupParent", []int64{1, 5000, 5002, 6000, 6002})
		assertTableColumn(t, db, table, "idGroupChild", []int64{5000, 5002, 6000, 6002, 7000})
	}
	assertTableColumn(t, db, "groups_items_propagate", "ID", []int64{1, 3, 4, 6})
	for _, table := range []string{
		"groups_items", "history_groups_items", "groups_attempts", "history_groups_attempts",
		"groups_login_prefixes", "history_groups_login_prefixes",
	} {
		assertTableColumn(t, db, table, "idGroup", []int64{5000, 5002, 6000, 6002})
	}
	assertTableColumn(t, db, "sessions", "idUser", []int64{500})
	for _, table := range []string{
		"users_threads", "history_users_threads", "users_answers", "users_items", "history_users_items",
		"filters", "history_filters", "refresh_tokens",
	} {
		assertTableColumn(t, db, table, "idUser", []int64{500, 502})
	}

	assertTableColumn(t, db, "groups_propagate", "sAncestorsComputationState", []string{"done"})
	found, err := store.GroupAncestors().Where("idGroupAncestor = 1 AND idGroupChild = 7000").HasRows()
	assert.NoError(t, err)
	assert.True(t, found, "No row for 1->7000 in groups_ancestors")
}

func setupDBForDeleteWithTrapsTests(t *testing.T, currentTime time.Time) *database.DB {
	db := testhelpers.SetupDBWithFixtureString(`
			groups_propagate: [{ID: 5000}, {ID: 5001}, {ID: 5002}, {ID: 6000}, {ID: 6001}, {ID: 6002}]
			history_users: [{ID: 500, iVersion: 1}, {ID: 501, iVersion: 1}, {ID: 502, iVersion: 1}]
			`, `
			users:
				- {ID: 500, tempUser: 1, sLogin: 500, idGroupSelf: 5000, idGroupOwned: 6000} # should be deleted
				- {ID: 501, sLogin: 501, tempUser: 1, idGroupSelf: 5001, idGroupOwned: 6001}
				- {ID: 502, sLogin: 502, idGroupSelf: 5002, idGroupOwned: 6002}
			sessions:
				- {idUser: 500, sExpirationDate: "`+currentTime.Format("2006-01-02 15:04:05")+`"}
				- {idUser: 501, sExpirationDate: "`+currentTime.Add(1*time.Second).Format("2006-01-02 15:04:05")+`"}
				- {idUser: 501, sExpirationDate: "`+currentTime.Add(-10*time.Second).Format("2006-01-02 15:04:05")+`"}
			users_threads: [{idUser: 500, idThread: 1}, {idUser: 501, idThread: 1}, {idUser: 502, idThread: 1}]
			history_users_threads: [{idUser: 500, ID: 1, idThread: 1, iVersion: 1}, {idUser: 501, ID: 2, idThread: 2, iVersion: 1},
			                        {idUser: 502, ID: 3, idThread: 3, iVersion: 1}]
			users_answers:
				- {idUser: 500, idItem: 1, sSubmissionDate: 2019-05-30 11:00:00}
				- {idUser: 501, idItem: 1, sSubmissionDate: 2019-05-30 11:00:00}
				- {idUser: 502, idItem: 1, sSubmissionDate: 2019-05-30 11:00:00}
			users_items: [{idUser: 500, idItem: 1}, {idUser: 501, idItem: 1}, {idUser: 502, idItem: 1}]
			history_users_items:
				- {idUser: 500, idItem: 1, ID: 1, nbSubmissionsAttempts: 0, nbTasksTried: 0, nbChildrenValidated: 0,
				   bValidated: 0, bFinished: 0, nbTasksWithHelp: 0, nbHintsCached: 0, nbCorrectionsRead: 0, iPrecision: 0,
				   iAutonomy: 0, bRanked: 0, iVersion: 1}
				- {idUser: 501, idItem: 1, ID: 2, nbSubmissionsAttempts: 0, nbTasksTried: 0, nbChildrenValidated: 0,
				   bValidated: 0, bFinished: 0, nbTasksWithHelp: 0, nbHintsCached: 0, nbCorrectionsRead: 0,
				   iPrecision: 0, iAutonomy: 0, bRanked: 0, iVersion: 1}
				- {idUser: 502, idItem: 1, ID: 3, nbSubmissionsAttempts: 0, nbTasksTried: 0, nbChildrenValidated: 0,
				   bValidated: 0, bFinished: 0, nbTasksWithHelp: 0, nbHintsCached: 0, nbCorrectionsRead: 0, iPrecision: 0,
				   iAutonomy: 0, bRanked: 0, iVersion: 1}
			filters: [{idUser: 500}, {idUser: 501}, {idUser: 502}]
			history_filters: [{idUser: 500, ID: 1, iVersion: 1}, {idUser: 501, ID: 2, iVersion: 1}, {idUser: 502, ID: 3, iVersion: 1}]
			refresh_tokens: [{idUser: 500, sRefreshToken: token}, {idUser: 501, sRefreshToken: token2},
			                 {idUser: 502, sRefreshToken: token3}]
			groups_items_propagate: [{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}]
			groups_items:
				- {ID: 1, idGroup: 5000, idItem: 1, idUserCreated: 1}
				- {ID: 2, idGroup: 5001, idItem: 1, idUserCreated: 1}
				- {ID: 3, idGroup: 5002, idItem: 1, idUserCreated: 1}
				- {ID: 4, idGroup: 6000, idItem: 1, idUserCreated: 1}
				- {ID: 5, idGroup: 6001, idItem: 1, idUserCreated: 1}
				- {ID: 6, idGroup: 6002, idItem: 1, idUserCreated: 1}
			history_groups_items:
				- {ID: 1, idGroup: 5000, idItem: 1, idUserCreated: 1, iVersion: 1}
				- {ID: 2, idGroup: 5001, idItem: 1, idUserCreated: 1, iVersion: 1}
				- {ID: 3, idGroup: 5002, idItem: 1, idUserCreated: 1, iVersion: 1}
				- {ID: 4, idGroup: 6000, idItem: 1, idUserCreated: 1, iVersion: 1}
				- {ID: 5, idGroup: 6001, idItem: 1, idUserCreated: 1, iVersion: 1}
				- {ID: 6, idGroup: 6002, idItem: 1, idUserCreated: 1, iVersion: 1}
			groups_attempts: [{idGroup: 5000, idItem: 1, iOrder: 0}, {idGroup: 5001, idItem: 1, iOrder: 0},
			                  {idGroup: 5002, idItem: 1, iOrder: 0}, {idGroup: 6000, idItem: 1, iOrder: 0},
			                  {idGroup: 6001, idItem: 1, iOrder: 0}, {idGroup: 6002, idItem: 1, iOrder: 0}]
			history_groups_attempts:
				- {idGroup: 5000, idItem: 1, ID: 1, iOrder: 0, iVersion: 1}
				- {idGroup: 5001, idItem: 1, ID: 2, iOrder: 0, iVersion: 1}
				- {idGroup: 5002, idItem: 1, ID: 3, iOrder: 0, iVersion: 1}
				- {idGroup: 6000, idItem: 1, ID: 4, iOrder: 0, iVersion: 1}
				- {idGroup: 6001, idItem: 1, ID: 5, iOrder: 0, iVersion: 1}
				- {idGroup: 6002, idItem: 1, ID: 6, iOrder: 0, iVersion: 1}
			groups_login_prefixes: [{idGroup: 5000, prefix: 5000, ID: 1}, {idGroup: 5001, prefix: 5001, ID: 2},
			                        {idGroup: 5002, prefix: 5002, ID: 3}, {idGroup: 6000, prefix: 6000, ID: 4},
			                        {idGroup: 6001, prefix: 6001, ID: 5}, {idGroup: 6002, prefix: 6002, ID: 6}]
			history_groups_login_prefixes:
				- {idGroup: 5000, prefix: 5000, ID: 1, iVersion: 1}
				- {idGroup: 5001, prefix: 5001, ID: 2, iVersion: 1}
				- {idGroup: 5002, prefix: 5002, ID: 3, iVersion: 1}
				- {idGroup: 6000, prefix: 5003, ID: 4, iVersion: 1}
				- {idGroup: 6001, prefix: 5004, ID: 5, iVersion: 1}
				- {idGroup: 6002, prefix: 5004, ID: 6, iVersion: 1}
			groups_groups: [{idGroupParent: 5000, idGroupChild: 7000}, {idGroupParent: 5001, idGroupChild: 7000},
			                {idGroupParent: 5002, idGroupChild: 7000}, {idGroupParent: 6000, idGroupChild: 7000},
			                {idGroupParent: 6001, idGroupChild: 7000}, {idGroupParent: 6002, idGroupChild: 7000},
			                {idGroupParent: 1, idGroupChild: 5000}, {idGroupParent: 1, idGroupChild: 5001},
			                {idGroupParent: 1, idGroupChild: 5002}, {idGroupParent: 1, idGroupChild: 6000},
			                {idGroupParent: 1, idGroupChild: 6001}, {idGroupParent: 1, idGroupChild: 6002}]
			history_groups_groups:
				- {idGroupParent: 5000, idGroupChild: 7000, ID: 1, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 5001, idGroupChild: 7000, ID: 2, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 5002, idGroupChild: 7000, ID: 3, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 6000, idGroupChild: 7000, ID: 4, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 6001, idGroupChild: 7000, ID: 5, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 6002, idGroupChild: 7000, ID: 6, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 5000, ID: 7, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 5001, ID: 8, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 5002, ID: 9, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 6000, ID: 10, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 6001, ID: 11, iChildOrder: 0, iVersion: 1}
				- {idGroupParent: 1, idGroupChild: 6002, ID: 12, iChildOrder: 0, iVersion: 1}
			groups_ancestors: [{idGroupAncestor: 1, idGroupChild: 1},
			                   {idGroupAncestor: 5000, idGroupChild: 7000}, {idGroupAncestor: 5001, idGroupChild: 7000},
			                   {idGroupAncestor: 5002, idGroupChild: 7000}, {idGroupAncestor: 6000, idGroupChild: 7000},
			                   {idGroupAncestor: 6001, idGroupChild: 7000}, {idGroupAncestor: 6002, idGroupChild: 7000},
			                   {idGroupAncestor: 1, idGroupChild: 5000}, {idGroupAncestor: 1, idGroupChild: 5001},
			                   {idGroupAncestor: 1, idGroupChild: 5002}, {idGroupAncestor: 1, idGroupChild: 6000},
			                   {idGroupAncestor: 1, idGroupChild: 6001}, {idGroupAncestor: 1, idGroupChild: 6002},
			                   {idGroupAncestor: 1, idGroupChild: 7000}]
			history_groups_ancestors:
				- {idGroupAncestor: 5000, idGroupChild: 7000, ID: 1, iVersion: 1}
				- {idGroupAncestor: 5001, idGroupChild: 7000, ID: 2, iVersion: 1}
				- {idGroupAncestor: 5002, idGroupChild: 7000, ID: 3, iVersion: 1}
				- {idGroupAncestor: 6000, idGroupChild: 7000, ID: 4, iVersion: 1}
				- {idGroupAncestor: 6001, idGroupChild: 7000, ID: 5, iVersion: 1}
				- {idGroupAncestor: 6002, idGroupChild: 7000, ID: 6, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 5000, ID: 7, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 5001, ID: 8, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 5002, ID: 8, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 6000, ID: 9, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 6001, ID: 10, iVersion: 1}
				- {idGroupAncestor: 1, idGroupChild: 6002, ID: 11, iVersion: 1}
			groups: [{ID: 1}, {ID: 5000}, {ID: 5001}, {ID: 5002}, {ID: 6000}, {ID: 6001}, {ID: 6002}, {ID: 7000}]
			history_groups:
				- {ID: 5000, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 5001, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 5002, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 6000, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 6001, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 6002, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}
				- {ID: 7000, bOpened: 0, bFreeAccess: 1, bSendEmails: 0, iVersion: 1}`)
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
