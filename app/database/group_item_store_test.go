package database

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
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
