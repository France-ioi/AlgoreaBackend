package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestGroupStore_ManagedBy(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, DefaultLanguage: "fr"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` " +
		"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id " +
		"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id " +
		"JOIN groups_ancestors_active AS user_ancestors " +
		"ON user_ancestors.ancestor_group_id = group_managers.manager_id AND " +
		"user_ancestors.child_group_id = ?")).
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().ManagedBy(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamGroupForUser(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, DefaultLanguage: "fr"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id AND "+
		"groups_groups_active.child_group_id = ? "+
		"WHERE (groups.id = ?) AND (groups.type = 'Team')")).
		WithArgs(int64(2), int64(4)).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupForUser(4, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_CreateNew_MustBeRunInTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupStore := NewDataStore(db).Groups()
	assert.PanicsWithValue(t, ErrNoTransaction,
		func() { _, _ = groupStore.CreateNew("", "") })
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_DeleteGroup_MustBeRunInTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupStore := NewDataStore(db).Groups()
	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = groupStore.DeleteGroup(1)
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_DeleteGroup_HandlesErrorOfInnerMethod(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery("^SELECT").WithArgs(int64(1234)).WillReturnError(expectedError)
	mock.ExpectRollback()

	store := NewDataStore(db)
	err := store.InTransaction(func(store *DataStore) error {
		return store.Groups().DeleteGroup(1234)
	})
	assert.Equal(t, expectedError, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_GroupIsVisibleToHimself(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()

	store := NewDataStore(db)

	assert.True(t, store.Groups().IsVisibleForGroup(1, 1))
}
