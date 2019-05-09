package database

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAnswerStore_WithMethods(t *testing.T) {
	tests := []struct {
		name          string
		expectedQuery string
	}{
		{
			name:          "WithUsers",
			expectedQuery: "SELECT `users_answers`.* FROM `users_answers` JOIN users ON users.ID = users_answers.idUser",
		},
		{
			name: "WithGroupAttempts",
			expectedQuery: "SELECT `users_answers`.* FROM `users_answers` " +
				"JOIN groups_attempts ON groups_attempts.ID = users_answers.idAttempt",
		},
		{
			name:          "WithItems",
			expectedQuery: "SELECT `users_answers`.* FROM `users_answers` JOIN items ON items.ID = users_answers.idItem",
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()

			mock.ExpectQuery("^" + regexp.QuoteMeta(testCase.expectedQuery) + "$").
				WillReturnRows(mock.NewRows([]string{"ID"}))

			store := NewDataStore(db).UserAnswers()
			resultValue := reflect.ValueOf(store).MethodByName(testCase.name).Call([]reflect.Value{})[0]
			newStore := resultValue.Interface().(*UserAnswerStore)

			assert.NotEqual(t, store, newStore)
			assert.Equal(t, "users_answers", newStore.DataStore.tableName)

			var result []interface{}
			err := newStore.Scan(&result).Error()
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
