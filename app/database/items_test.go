package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestDB_JoinsUserAndDefaultItemStrings(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := auth.NewMockUser(1, 2, 3, 4)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` LEFT JOIN items_strings default_strings FORCE INDEX (idItem) ON default_strings.idItem = items.ID AND default_strings.idLanguage = items.idDefaultLanguage LEFT JOIN items_strings user_strings ON user_strings.idItem=items.ID AND user_strings.idLanguage = ?")).
		WithArgs(4).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := db.Table("items").JoinsUserAndDefaultItemStrings(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
