package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGroupAncestorStore_OwnedByUser(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 11, DefaultLanguageID: 0})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupAncestor=?")).
		WithArgs(11).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).GroupAncestors().OwnedByUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupAncestorStore_OwnedByUser_HandlesUserError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(123).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	user := database.NewUser(123, database.NewDataStore(db).Users(), nil)

	var result []interface{}
	err := database.NewDataStore(db).GroupAncestors().OwnedByUser(user).Scan(&result).Error()
	assert.Equal(t, database.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupAncestorStore_UserAncestors_HandlesUserError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(123).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	user := database.NewUser(123, database.NewDataStore(db).Users(), nil)

	var result []interface{}
	err := database.NewDataStore(db).GroupAncestors().UserAncestors(user).Scan(&result).Error()
	assert.Equal(t, database.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
