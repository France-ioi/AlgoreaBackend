package database

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT 1 FROM `groups_ancestors`  "+
			"WHERE (child_group_id = ? AND ancestor_group_id = ?) LIMIT 1 FOR UPDATE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs("groups_groups").
		WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectRollback()

	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.GroupGroups().CreateRelation(1, 2)
	})
	assert.Equal(t, ErrRelationCycle, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelationsWithoutChecking_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupGroupStore := NewDataStore(db).GroupGroups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = groupGroupStore.CreateRelationsWithoutChecking([]map[string]interface{}{{"parent_group_id": 1, "child_group_id": 2}})
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
	)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `parent_group_id`) "+
			"VALUES (?, ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.inTransaction(func(db *DB) error {
		groupGroupStore := NewDataStore(db).GroupGroups()
		groupGroupStore.createRelation(1, 2)
		return nil
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
