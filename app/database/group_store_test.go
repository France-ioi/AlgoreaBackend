package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupStore_ManagedBy(t *testing.T) {
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

func TestGroupStore_TeamGroupForTeamItemAndUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, DefaultLanguage: "fr"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id AND "+
		"groups_groups_active.child_group_id = ? "+
		"JOIN attempts ON attempts.participant_id = groups.id AND attempts.root_item_id = ? AND "+
		"NOW() < attempts.allows_submissions_until "+
		"WHERE (groups.type = 'Team') ORDER BY `groups`.`id` LIMIT 1")).
		WithArgs(2, 1234).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupForTeamItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamGroupForUser(t *testing.T) {
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
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupStore := NewDataStore(db).Groups()
	assert.PanicsWithValue(t, ErrNoTransaction,
		func() { _, _ = groupStore.CreateNew("", "", nil) })
	assert.NoError(t, mock.ExpectationsWereMet())
}
