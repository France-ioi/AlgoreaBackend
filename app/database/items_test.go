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

	mockUser := &database.User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

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

	mockUser := &database.User{GroupID: 2, OwnedGroupID: ptrInt64(3), DefaultLanguageID: 4}

	mock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT `items`.* FROM `items` JOIN (SELECT item_id, MAX(can_view_generated_value) AS can_view_generated_value, "+
			"MAX(can_grant_view_generated_value) AS can_grant_view_generated_value, "+
			"MAX(can_watch_generated_value) AS can_watch_generated_value, "+
			"MAX(can_edit_generated_value) AS can_edit_generated_value, "+
			"MAX(is_owner_generated) AS is_owner_generated "+
			"FROM permissions_generated AS permissions JOIN ( "+
			"SELECT * FROM groups_ancestors_active "+
			"WHERE groups_ancestors_active.child_group_id = ? "+
			") AS ancestors "+
			"ON ancestors.ancestor_group_id = permissions.group_id "+
			"WHERE (can_view_generated_value >= ?) "+
			"GROUP BY permissions.item_id) "+
			"as visible ON visible.item_id = items.id")+"$").
		WithArgs(2, database.NewDataStore(db).PermissionsGranted().ViewIndexByName("info")).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := db.Table("items").WhereItemsAreVisible(mockUser).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func ptrInt64(i int64) *int64 { return &i }
