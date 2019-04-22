// +build !unit

package database_test

import (
	"regexp"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func setupDB() *database.DB {
	return testhelpers.SetupDBWithFixture("visibility")
}

func TestVisible(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := database.NewMockUser(1, &database.UserData{SelfGroupID: 11, OwnedGroupID: 12, DefaultLanguageID: 2})
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []int64
	db = itemStore.Visible(user).Pluck("ID", &result)
	assert.NoError(db.Error())

	expected := []int64{190, 191, 192, 1900, 1901, 1902, 19000, 19001, 19002}
	assert.Equal(expected, result)
}

func TestVisibleByID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := database.NewMockUser(1, &database.UserData{SelfGroupID: 11, OwnedGroupID: 12, DefaultLanguageID: 2})
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []int64
	db = itemStore.VisibleByID(user, 191).Pluck("ID", &result)
	assert.NoError(db.Error())

	expected := []int64{191}
	assert.Equal(expected, result)
}

func TestVisibleChildrenOfID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := database.NewMockUser(1, &database.UserData{SelfGroupID: 11, OwnedGroupID: 12, DefaultLanguageID: 2})
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []int64
	db = itemStore.VisibleChildrenOfID(user, 190).Pluck("items.ID", &result)
	assert.NoError(db.Error())

	expected := []int64{1900, 1901, 1902}
	assert.Equal(expected, result)
}

func TestVisibleGrandChildrenOfID(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB()
	defer func() { _ = db.Close() }()
	user := database.NewMockUser(1, &database.UserData{SelfGroupID: 11, OwnedGroupID: 12, DefaultLanguageID: 2})
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []int64
	db = itemStore.VisibleGrandChildrenOfID(user, 190).Pluck("items.ID", &result)
	assert.NoError(db.Error())

	expected := []int64{19000, 19001, 19002}
	assert.Equal(expected, result)
}

func TestItemStore_AccessRights(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()

	mockUser := database.NewMockUser(1, &database.UserData{SelfGroupID: 2, OwnedGroupID: 3, DefaultLanguageID: 4})

	mock.ExpectQuery("^" + regexp.QuoteMeta(
		"SELECT idItem, MIN(sCachedFullAccessDate) <= NOW() AS fullAccess, "+
			"MIN(sCachedPartialAccessDate) <= NOW() AS partialAccess, "+
			"MIN(sCachedGrayedAccessDate) <= NOW() AS grayedAccess, "+
			"MIN(sCachedAccessSolutionsDate) <= NOW() AS accessSolutions "+
			"FROM `groups_items` "+
			"JOIN (SELECT * FROM `groups_ancestors` WHERE (groups_ancestors.idGroupChild = ?)) AS ancestors "+
			"ON groups_items.idGroup = ancestors.idGroupAncestor GROUP BY idItem") + "$").
		WithArgs(2).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := database.NewDataStore(db).Items().AccessRights(mockUser).Scan(&result).Error()
	assertlib.NoError(t, err)
	assertlib.NoError(t, mock.ExpectationsWereMet())
}
