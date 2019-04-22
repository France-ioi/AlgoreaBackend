package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAnswerStore_WithUsers(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `users_answers`.* FROM `users_answers` JOIN users ON users.ID = users_answers.idUser")).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	store := NewDataStore(db).UserAnswers()
	newStore := store.WithUsers()
	assert.NotEqual(t, store, newStore)
	assert.Equal(t, "users_answers", newStore.DataStore.tableName)

	var result []interface{}
	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAnswerStore_WithGroupAttempts(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `users_answers`.* FROM `users_answers` " +
		"JOIN groups_attempts ON groups_attempts.ID = users_answers.idAttempt")).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	store := NewDataStore(db).UserAnswers()
	newStore := store.WithGroupAttempts()
	assert.NotEqual(t, store, newStore)
	assert.Equal(t, "users_answers", newStore.DataStore.tableName)

	var result []interface{}
	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAnswerStore_WithItems(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `users_answers`.* FROM `users_answers` JOIN items ON items.ID = users_answers.idItem")).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	store := NewDataStore(db).UserAnswers()
	newStore := store.WithItems()
	assert.NotEqual(t, store, newStore)
	assert.Equal(t, "users_answers", newStore.DataStore.tableName)

	var result []interface{}
	err := newStore.Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
