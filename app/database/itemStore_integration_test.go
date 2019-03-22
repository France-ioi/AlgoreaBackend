// +build !unit

package database_test

import (
	"regexp"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func setupDB() *database.DB {
	return testhelpers.SetupDBWithFixture("visibility")
}

type itemIdRow struct {
	ID int `sql:"column:ID"`
}

func TestVisible(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11, 12, 2)
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []itemIdRow
	db = itemStore.Visible(user).Select("ID").Scan(&result)
	assert.NoError(db.Error())

	expected := []itemIdRow{{ID: 190}, {ID: 191}, {ID: 192}, {ID: 1900}, {ID: 1901}, {ID: 1902}, {ID: 19000}, {ID: 19001}, {ID: 19002}}
	assert.Equal(expected, result)
}

func TestVisibleByID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11, 12, 2)
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []itemIdRow
	db = itemStore.VisibleByID(user, 191).Select("ID").Scan(&result)
	assert.NoError(db.Error())

	expected := []itemIdRow{{ID: 191}}
	assert.Equal(expected, result)
}

func TestVisibleChildrenOfID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11, 12, 2)
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []itemIdRow
	db = itemStore.VisibleChildrenOfID(user, 190).Select("items.ID").Scan(&result)
	assert.NoError(db.Error())

	expected := []itemIdRow{{ID: 1900}, {ID: 1901}, {ID: 1902}}
	assert.Equal(expected, result)
}

func TestVisibleGrandChildrenOfID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11, 12, 2)
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []itemIdRow
	db = itemStore.VisibleGrandChildrenOfID(user, 190).Select("items.ID").Scan(&result)
	assert.NoError(db.Error())

	expected := []itemIdRow{{ID: 19000}, {ID: 19001}, {ID: 19002}}
	assert.Equal(expected, result)
}

func TestItemStore_AccessRights(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := auth.NewMockUser(1, 2, 3, 4)

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions FROM `groups_items` JOIN (SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors ON groups_items.idGroup = ancestors.idGroupAncestor GROUP BY idItem") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).Items().AccessRights(mockUser).Scan(&result).Error()
	assertlib.NoError(t, err)
	assertlib.NoError(t, mock.ExpectationsWereMet())
}
