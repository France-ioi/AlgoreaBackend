package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupStore_ManagedBy(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` " +
		"JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups.id " +
		"JOIN group_managers ON group_managers.group_id = groups_ancestors_active.ancestor_group_id " +
		"AND group_managers.manager_id = ?")).
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

	mockUser := &User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id AND "+
		"groups_groups_active.child_group_id = ? "+
		"WHERE (groups.team_item_id = ?) AND (groups.type = 'Team') ORDER BY `groups`.`id` LIMIT 1")).
		WithArgs(2, 1234).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupForTeamItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamGroupForItemAndUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` "+
		"JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id AND "+
		"groups_groups_active.child_group_id = ? "+
		"LEFT JOIN items_ancestors ON items_ancestors.ancestor_item_id = groups.team_item_id "+
		"WHERE (groups.type = 'Team') AND (items_ancestors.child_item_id = ? OR groups.team_item_id = ?) "+
		"GROUP BY groups.id ORDER BY `groups`.`id` LIMIT 1")).
		WithArgs(2, 1234, 1234).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamGroupForItemAndUser(1234, mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_TeamsMembersForItem(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` JOIN groups_groups_active "+
		"ON groups_groups_active.parent_group_id = groups.id"+
		" WHERE (groups.type = 'Team') AND (groups_groups_active.child_group_id IN (?,?,?)) AND (groups.team_item_id = ?)")).
		WithArgs(1, 2, 3, 1234).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStore(db).Groups().TeamsMembersForItem([]int64{1, 2, 3}, 1234).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_CreateNew_MustBeRunInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	groupStore := NewDataStore(db).Groups()
	assert.PanicsWithValue(t, ErrNoTransaction,
		func() { _, _ = groupStore.CreateNew(nil, nil, nil) })
	assert.NoError(t, mock.ExpectationsWereMet())
}
