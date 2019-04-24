package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupAncestorStore_OwnedByUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := NewMockUser(1, &UserData{SelfGroupID: 2, OwnedGroupID: 11, DefaultLanguageID: 0})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupAncestor=?")).
		WithArgs(11).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).GroupAncestors().OwnedByUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupAncestorStore_OwnedByUser_HandlesUserError(t *testing.T) {
	testMethodHandlesUserNotFoundError(t, func(db *DB, user *User) []interface{} {
		var result []interface{}
		err := NewDataStore(db).GroupAncestors().OwnedByUser(user).Scan(&result).Error()
		return []interface{}{err}
	}, []interface{}{ErrUserNotFound})
}

func TestGroupAncestorStore_UserAncestors_HandlesUserError(t *testing.T) {
	testMethodHandlesUserNotFoundError(t, func(db *DB, user *User) []interface{} {
		var result []interface{}
		err := NewDataStore(db).GroupAncestors().UserAncestors(user).Scan(&result).Error()
		return []interface{}{err}
	}, []interface{}{ErrUserNotFound})
}
