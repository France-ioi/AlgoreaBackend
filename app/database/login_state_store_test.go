package database

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginStateStore_DeleteExpired(t *testing.T) {
	db, mock := NewDBMock()
	expectedError := errors.New("some error")
	mock.ExpectExec("^" + regexp.QuoteMeta("DELETE FROM `login_states` WHERE (expiration_date <= NOW())")).
		WillReturnError(expectedError)

	err := NewDataStore(db).LoginStates().DeleteExpired()

	assert.Equal(t, expectedError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
