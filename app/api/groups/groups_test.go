package groups

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func TestService_checkThatUserOwnsTheGroup_HandlesError(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(1).
		WillReturnError(expectedError)

	user := database.NewUser(1, database.NewDataStore(db).Users(), nil)

	apiErr := checkThatUserOwnsTheGroup(database.NewDataStore(db), user, 123)

	assert.Equal(t, service.ErrUnexpected(expectedError), apiErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkThatUserHasRightsForDirectRelation_UserNotFound(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "idDefaultLanguage"}))
	mock.ExpectCommit()

	user := database.NewUser(1, database.NewDataStore(db).Users(), nil)
	assert.NoError(t, database.NewDataStore(db).InTransaction(func(store *database.DataStore) error {
		got := checkThatUserHasRightsForDirectRelation(store, user, 1, 2)
		assert.Equal(t, got, service.InsufficientAccessRightsError)
		return nil
	}))

	assert.NoError(t, mock.ExpectationsWereMet())
}
