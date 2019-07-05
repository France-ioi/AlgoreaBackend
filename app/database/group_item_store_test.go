package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGroupItemStore_MatchingUserAncestors_HandlesError(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(1).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	user := NewUser(1, NewDataStore(db).Users(), nil)
	var result []interface{}
	err := NewDataStore(db).GroupItems().MatchingUserAncestors(user).Scan(&result).Error()
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupItemStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).GroupItems().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupItemStore_After_HandlesErrorOfComputeAllAccess(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectPrepare("^INSERT IGNORE INTO groups_items").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).GroupItems().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestItemItemStore_After_HandlesErrorOfGrantCachedAccessWhereNeeded(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	expectedComputeAllAccessSQL := []string{
		"^INSERT IGNORE INTO groups_items",
		"^INSERT INTO groups_items_propagate",
		"^UPDATE groups_items",
		"^INSERT INTO groups_items_propagate",
		"^INSERT INTO groups_items_propagate",
		"^UPDATE groups_items_propagate",
		"^UPDATE groups_items",
		"^UPDATE groups_items",
		"^UPDATE groups_items",
		"^UPDATE groups_items",
		"^UPDATE groups_items",
		"^UPDATE groups_items_propagate",
	}
	for _, sql := range expectedComputeAllAccessSQL {
		dbMock.ExpectPrepare(sql)
	}
	for _, sql := range expectedComputeAllAccessSQL {
		dbMock.ExpectExec(sql).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	dbMock.ExpectExec("^UPDATE `groups_items`").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).GroupItems().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGroupItemStore_AccessRightsForItemsVisibleToUser(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := NewMockUser(1, &UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
			"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, "+
			"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions "+
			"FROM `groups_items` JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors "+
			"ON ancestors.idGroupAncestor = groups_items.idGroup GROUP BY groups_items.idItem "+
			"HAVING (fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0)") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).GroupItems().AccessRightsForItemsVisibleToUser(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
