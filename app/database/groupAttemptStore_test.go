package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupAttemptStore_ByID(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const attemptID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `groups_attempts` WHERE (groups_attempts.ID = ?)")).
		WithArgs(attemptID).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).GroupAttempts().ByID(attemptID).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
