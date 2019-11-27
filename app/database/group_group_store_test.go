package database

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestGroupGroupStore_WhereUserIsMember(t *testing.T) {
	for _, test := range []struct {
		tableName     string
		expectedQuery string
		storeFunc     func(*DB) *GroupGroupStore
	}{
		{
			tableName: "groups_groups",
			expectedQuery: "SELECT * FROM `groups_groups` " +
				"WHERE (`groups_groups`.child_group_id = ?) AND (NOW() < groups_groups.expires_at)",
			storeFunc: func(db *DB) *GroupGroupStore { return NewDataStore(db).GroupGroups() },
		},
		{
			tableName: "groups_groups_active",
			expectedQuery: "SELECT * FROM `groups_groups_active` " +
				"WHERE (`groups_groups_active`.child_group_id = ?)",
			storeFunc: func(db *DB) *GroupGroupStore { return NewDataStore(db).ActiveGroupGroups() },
		},
	} {
		test := test
		t.Run(test.tableName, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			mockUser := &User{GroupID: 2}
			mock.ExpectQuery(regexp.QuoteMeta(test.expectedQuery)).
				WithArgs(2).
				WillReturnRows(mock.NewRows([]string{"id"}))

			var result []interface{}
			err := test.storeFunc(db).WhereUserIsMember(mockUser).Scan(&result).Error()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGroupGroupStore_CreateRelation(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const (
		parentGroupID = 1
		childGroupID  = 2
	)

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(1)))
	mock.ExpectExec("^"+
		regexp.QuoteMeta("DELETE FROM `groups_groups`  "+
			"WHERE (child_group_id = ? AND parent_group_id = ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT id FROM `groups_ancestors`  "+
			"WHERE (child_group_id = ? AND ancestor_group_id = ?) LIMIT 1 FOR UPDATE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("^" +
		regexp.QuoteMeta("SET @maxIChildOrder = IFNULL((SELECT MAX(child_order) FROM `groups_groups` "+
			"WHERE `parent_group_id` = ? FOR UPDATE), 0)") + "$").
		WithArgs(parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 0))

	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`) "+
			"VALUES (?, @maxIChildOrder+1, ?, ?)")+"$").
		WithArgs(childGroupID, sqlmock.AnyArg(), parentGroupID).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'PRIMARY'"})
	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`) "+
			"VALUES (?, @maxIChildOrder+1, ?, ?)")+"$").
		WithArgs(childGroupID, sqlmock.AnyArg(), parentGroupID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	setMockExpectationsForCreateNewAncestors(mock)

	mock.ExpectExec("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs("groups_groups").
		WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectCommit()

	err := db.inTransaction(func(db *DB) error {
		groupGroupStore := NewDataStore(db).GroupGroups()
		return groupGroupStore.CreateRelation(1, 2)
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelation_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupGroupStore := NewDataStore(db).GroupGroups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() { _ = groupGroupStore.CreateRelation(1, 2) })
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelation_PreventsRelationCycles(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const (
		parentGroupID = 1
		childGroupID  = 2
	)

	mock.ExpectBegin()
	mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(1)))
	mock.ExpectExec("^"+
		regexp.QuoteMeta("DELETE FROM `groups_groups`  "+
			"WHERE (child_group_id = ? AND parent_group_id = ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT id FROM `groups_ancestors`  "+
			"WHERE (child_group_id = ? AND ancestor_group_id = ?) LIMIT 1 FOR UPDATE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs("groups_groups").
		WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectRollback()

	err := db.inTransaction(func(db *DB) error {
		groupGroupStore := NewDataStore(db).GroupGroups()
		return groupGroupStore.CreateRelation(1, 2)
	})
	assert.Equal(t, ErrRelationCycle, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelationsWithoutChecking(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	relations := []ParentChild{{ParentID: 1, ChildID: 2}, {ParentID: 3, ChildID: 4}}

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(1)))

	for i, relation := range relations {
		mock.ExpectExec("^" +
			regexp.QuoteMeta("SET @maxIChildOrder = IFNULL((SELECT MAX(child_order) FROM `groups_groups` "+
				"WHERE `parent_group_id` = ? FOR UPDATE), 0)") + "$").
			WithArgs(relation.ParentID).
			WillReturnResult(sqlmock.NewResult(-1, 0))

		mock.ExpectExec("^"+
			regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`) "+
				"VALUES (?, @maxIChildOrder+1, ?, ?)")+"$").
			WithArgs(relation.ChildID, sqlmock.AnyArg(), relation.ParentID).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'PRIMARY'"})
		mock.ExpectExec("^"+
			regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`) "+
				"VALUES (?, @maxIChildOrder+1, ?, ?)")+"$").
			WithArgs(relation.ChildID, sqlmock.AnyArg(), relation.ParentID).
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
	}

	setMockExpectationsForCreateNewAncestors(mock)

	mock.ExpectExec("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs("groups_groups").
		WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectCommit()

	err := db.inTransaction(func(db *DB) error {
		groupGroupStore := NewDataStore(db).GroupGroups()
		return groupGroupStore.CreateRelationsWithoutChecking(relations)
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func setMockExpectationsForCreateNewAncestors(mock sqlmock.Sqlmock) {
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectPrepare("")
	mock.ExpectPrepare("")
	mock.ExpectPrepare("")
	mock.ExpectPrepare("")
	mock.ExpectPrepare("")
	mock.ExpectPrepare("")
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
}

func TestGroupGroupStore_CreateRelationsWithoutChecking_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupGroupStore := NewDataStore(db).GroupGroups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = groupGroupStore.CreateRelationsWithoutChecking([]ParentChild{{ParentID: 1, ChildID: 2}})
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_DeleteRelation_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupGroupStore := NewDataStore(db).GroupGroups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = groupGroupStore.DeleteRelation(1, 2, true)
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_DeleteRelation_ShouldUseNamedLock(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs("groups_groups", groupsRelationsLockTimeout/time.Second).
		WillReturnRows(sqlmock.NewRows([]string{"SELECT GET_LOCK(?, ?)"}).AddRow(int64(0)))
	mock.ExpectRollback()

	store := NewDataStore(db)
	_ = store.InTransaction(func(store *DataStore) error {
		return store.GroupGroups().DeleteRelation(1, 2, true)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_createRelation(t *testing.T) {
	const (
		parentGroupID = 1
		childGroupID  = 2
		role          = "role"
	)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectExec("^" +
		regexp.QuoteMeta("SET @maxIChildOrder = IFNULL((SELECT MAX(child_order) FROM `groups_groups` "+
			"WHERE `parent_group_id` = ? FOR UPDATE), 0)") + "$").
		WithArgs(parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 0))

	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`, `role`) "+
			"VALUES (?, @maxIChildOrder+1, ?, ?, ?)")+"$").
		WithArgs(childGroupID, sqlmock.AnyArg(), parentGroupID, role).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'PRIMARY'"})
	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `child_order`, `id`, `parent_group_id`, `role`) "+
			"VALUES (?, @maxIChildOrder+1, ?, ?, ?)")+"$").
		WithArgs(childGroupID, sqlmock.AnyArg(), parentGroupID, role).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.inTransaction(func(db *DB) error {
		groupGroupStore := NewDataStore(db).GroupGroups()
		groupGroupStore.createRelation(1, 2, role)
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).GroupGroups().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupGroupStore_After_HandlesErrorOfCreateNewAncestors(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectExec("^INSERT INTO  groups_propagate").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).GroupGroups().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}
