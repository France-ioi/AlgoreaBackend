package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestGroupGroupStore_WhereUserIsMember(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_groups` " +
		"WHERE (groups_groups.idGroupChild = ?) AND (groups_groups.sType" + GroupRelationIsActiveCondition + ")")).
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).GroupGroups().WhereUserIsMember(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
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
			"WHERE (idGroupChild = ? AND idGroupParent = ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT ID FROM `groups_ancestors`  "+
			"WHERE (idGroupChild = ? AND idGroupAncestor = ?) LIMIT 1 FOR UPDATE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))
	mock.ExpectExec("^" +
		regexp.QuoteMeta("SET @maxIChildOrder = IFNULL((SELECT MAX(iChildOrder) FROM `groups_groups` "+
			"WHERE `idGroupParent` = ? FOR UPDATE), 0)") + "$").
		WithArgs(parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 0))

	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) "+
			"VALUES (?, ?, ?, @maxIChildOrder+1)")+"$").
		WithArgs(sqlmock.AnyArg(), parentGroupID, childGroupID).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'PRIMARY'"})
	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) "+
			"VALUES (?, ?, ?, @maxIChildOrder+1)")+"$").
		WithArgs(sqlmock.AnyArg(), parentGroupID, childGroupID).
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
			"WHERE (idGroupChild = ? AND idGroupParent = ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT ID FROM `groups_ancestors`  "+
			"WHERE (idGroupChild = ? AND idGroupAncestor = ?) LIMIT 1 FOR UPDATE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(int64(1)))
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
			regexp.QuoteMeta("SET @maxIChildOrder = IFNULL((SELECT MAX(iChildOrder) FROM `groups_groups` "+
				"WHERE `idGroupParent` = ? FOR UPDATE), 0)") + "$").
			WithArgs(relation.ParentID).
			WillReturnResult(sqlmock.NewResult(-1, 0))

		mock.ExpectExec("^"+
			regexp.QuoteMeta("INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) "+
				"VALUES (?, ?, ?, @maxIChildOrder+1)")+"$").
			WithArgs(sqlmock.AnyArg(), relation.ParentID, relation.ChildID).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '1' for key 'PRIMARY'"})
		mock.ExpectExec("^"+
			regexp.QuoteMeta("INSERT INTO groups_groups (ID, idGroupParent, idGroupChild, iChildOrder) "+
				"VALUES (?, ?, ?, @maxIChildOrder+1)")+"$").
			WithArgs(sqlmock.AnyArg(), relation.ParentID, relation.ChildID).
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
