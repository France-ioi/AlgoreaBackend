package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestGroupStore_OwnedBy(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `groups`.* FROM `groups` JOIN groups_ancestors ON groups_ancestors.idGroupChild = groups.ID WHERE (groups_ancestors.idGroupAncestor=?)")).
		WithArgs(3).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).Groups().OwnedBy(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupStore_OwnedBy_HandlesError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(1).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	user := database.NewUser(1, database.NewDataStore(db).Users(), nil)
	var result []interface{}
	err := database.NewDataStore(db).Groups().OwnedBy(user).Scan(&result).Error()
	assert.Equal(t, database.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
