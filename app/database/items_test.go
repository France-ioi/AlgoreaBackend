package database_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func TestDB_JoinsUserAndDefaultItemStrings(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &database.User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` LEFT JOIN items_strings default_strings FORCE INDEX (item_id) " +
			"ON default_strings.item_id = items.id AND default_strings.language_id = items.default_language_id " +
			"LEFT JOIN items_strings user_strings ON user_strings.item_id=items.id AND user_strings.language_id = ?")).
		WithArgs(4).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := db.Table("items").JoinsUserAndDefaultItemStrings(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_WhereItemsAreVisible(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := &database.User{ID: 1, SelfGroupID: ptrInt64(2), OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` JOIN (SELECT item_id, MIN(cached_full_access_date) <= NOW() AS full_access, "+
			"MIN(cached_partial_access_date) <= NOW() AS partial_access, MIN(cached_grayed_access_date) <= NOW() AS grayed_access, "+
			"MIN(cached_access_solutions_date) <= NOW() AS access_solutions "+
			"FROM `groups_items` JOIN (SELECT * FROM groups_ancestors WHERE (groups_ancestors.child_group_id = ?)) AS ancestors "+
			"ON ancestors.ancestor_group_id = groups_items.group_id GROUP BY groups_items.item_id "+
			"HAVING (full_access > 0 OR partial_access > 0 OR grayed_access > 0)) "+
			"as visible ON visible.item_id = items.id") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := db.Table("items").WhereItemsAreVisible(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func ptrInt64(i int64) *int64 { return &i }
