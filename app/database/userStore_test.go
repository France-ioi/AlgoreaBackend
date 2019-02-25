package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserStore_ByID(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const userID = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (users.ID = ?)")).
		WithArgs(userID).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStore(db).Users().ByID(userID).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
