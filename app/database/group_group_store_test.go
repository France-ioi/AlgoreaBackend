package database

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
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
			testoutput.SuppressIfPasses(t)

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
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	var resultsPropagationScheduled bool
	monkey.PatchInstanceMethod(reflect.TypeOf(&DataStore{}), "ScheduleResultsPropagation",
		func(_ *DataStore) { resultsPropagationScheduled = true })
	defer monkey.UnpatchAll()

	const (
		parentGroupID = 1
		childGroupID  = 2
	)

	fakeDBEnums("permissions_granted.can_view",
		map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5},
		map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"})
	defer ClearAllDBEnums()

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT 1 FROM `groups_ancestors`  "+
			"WHERE (child_group_id = ? AND ancestor_group_id = ?) LIMIT 1 FOR SHARE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectExec("^"+
		regexp.QuoteMeta("DELETE FROM `groups_groups`  "+
			"WHERE (child_group_id = ? AND parent_group_id = ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	mock.ExpectExec("^"+
		regexp.QuoteMeta("DELETE FROM `group_pending_requests`  "+
			"WHERE (group_id = ? AND member_id = ?)")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `parent_group_id`) "+
			"VALUES (?, ?)")+"$").
		WithArgs(childGroupID, parentGroupID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	setMockExpectationsForCreateNewAncestors(mock)

	mock.ExpectCommit()

	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.GroupGroups().CreateRelation(1, 2)
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.True(t, resultsPropagationScheduled)
}

func TestGroupGroupStore_CreateRelation_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupGroupStore := NewDataStore(db).GroupGroups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() { _ = groupGroupStore.CreateRelation(1, 2) })
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelation_PreventsRelationCycles(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const (
		parentGroupID = 1
		childGroupID  = 2
	)

	mock.ExpectBegin()
	mock.ExpectQuery("^"+
		regexp.QuoteMeta("SELECT 1 FROM `groups_ancestors`  "+
			"WHERE (child_group_id = ? AND ancestor_group_id = ?) LIMIT 1 FOR SHARE")+"$").
		WithArgs(parentGroupID, childGroupID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectRollback()

	err := NewDataStore(db).InTransaction(func(s *DataStore) error {
		return s.GroupGroups().CreateRelation(1, 2)
	})
	assert.Equal(t, ErrRelationCycle, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateRelationsWithoutChecking(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	relations := []map[string]interface{}{
		{"parent_group_id": int64(1), "child_group_id": int64(2)},
		{"parent_group_id": int64(3), "child_group_id": int64(4)},
	}

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()

	mock.ExpectExec("^"+
		regexp.QuoteMeta("INSERT INTO `groups_groups` (`child_group_id`, `parent_group_id`) "+
			"VALUES (?, ?), (?, ?)")+"$").
		WithArgs(int64(2), int64(1), int64(4), int64(3)).
		WillReturnResult(sqlmock.NewResult(int64(1), 1))

	mock.MatchExpectationsInOrder(false)
	setMockExpectationsForCreateNewAncestors(mock)

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
	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(-1, 0))
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

func TestGroupGroupStore_createRelation(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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

func TestGroupGroupStore_CreateNewAncestors_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).GroupGroups().CreateNewAncestors()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupGroupStore_CreateNewAncestors_HandlesErrorOfCreateNewAncestors(t *testing.T) {
	testErrorHandlingInAncestorsPropagation(t, "groups_propagate", func(trDB *DB) error {
		return NewDataStore(trDB).GroupGroups().CreateNewAncestors()
	})
}

func testErrorHandlingInAncestorsPropagation(t *testing.T, propagateTableName string, callPropagationFunc func(db *DB) error) {
	t.Helper()
	testoutput.SuppressIfPasses(t)

	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectExec("^INSERT INTO  " + propagateTableName).WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return callPropagationFunc(trDB)
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupGroupStore_TeamGroupForTeamItemAndUser(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, DefaultLanguage: "fr"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups_groups_active`.* FROM `groups_groups_active` "+
		"JOIN attempts ON attempts.participant_id = groups_groups_active.parent_group_id AND attempts.root_item_id = ? AND "+
		"NOW() < attempts.allows_submissions_until "+
		"WHERE (groups_groups_active.is_team_membership = 1) AND (groups_groups_active.child_group_id = ?) "+
		"ORDER BY groups_groups_active.parent_group_id LIMIT 1")).
		WithArgs(1234, 2).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).ActiveGroupGroups().TeamGroupForTeamItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
