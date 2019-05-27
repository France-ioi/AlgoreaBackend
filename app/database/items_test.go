package database_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestDB_JoinsUserAndDefaultItemStrings(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` LEFT JOIN items_strings default_strings FORCE INDEX (idItem) " +
			"ON default_strings.idItem = items.ID AND default_strings.idLanguage = items.idDefaultLanguage " +
			"LEFT JOIN items_strings user_strings ON user_strings.idItem=items.ID AND user_strings.idLanguage = ?")).
		WithArgs(4).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := db.Table("items").JoinsUserAndDefaultItemStrings(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_WhereItemsAreVisible(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` JOIN (SELECT idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
			"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, "+
			"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions "+
			"FROM `groups_items` JOIN (SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors "+
			"ON groups_items.idGroup = ancestors.idGroupAncestor GROUP BY idItem) as visible ON visible.idItem = items.ID "+
			"WHERE (fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0)") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := db.Table("items").WhereItemsAreVisible(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_ItemMethodsHandleUserError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "WhereItemsAreVisible"},
		{name: "JoinsUserAndDefaultItemStrings"},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			db, mock := database.NewDBMock()
			defer func() { _ = db.Close() }()

			mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT users.*, l.ID as idDefaultLanguage FROM `users`")).
				WithArgs(1).
				WillReturnRows(mock.NewRows([]string{"ID"}))

			user := database.NewUser(1, database.NewDataStore(db).Users(), nil)
			var result []interface{}
			err := reflect.ValueOf(db.Table("items")).MethodByName(testCase.name).
				Call([]reflect.Value{reflect.ValueOf(user)})[0].Interface().(*database.DB).
				Scan(&result).Error()
			assert.Equal(t, database.ErrUserNotFound, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
