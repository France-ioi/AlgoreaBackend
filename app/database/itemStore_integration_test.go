package database_test

import (
	"database/sql"
	"testing"

	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func setupDB(t *testing.T) *database.DB {
	var err error

	if testhelpers.HasNoDBFlag() {
		t.Skip("Skipping BDD tests in package 'api' (TESTS_NODB env set)")
	}

	// needs actual config for connection to DB
	var conf *config.Root
	if conf, err = config.Load(); err != nil {
		t.Fatal(err)
	}

	// Seed the DB
	var rawDb *sql.DB
	rawDb, err = sql.Open("mysql", conf.Database.Connection.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDb.Close() }()
	testhelpers.EmptyDB(t, rawDb, conf.Database.Connection.DBName)
	testhelpers.LoadFixture(t, rawDb, "visibility")

	// Return a new db connection
	var db *database.DB
	db, err = database.Open(conf.Database.Connection.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type itemIdRow struct {
	ID int `sql:"column:ID"`
}

func TestVisible(t *testing.T) {
	assert := assertlib.New(t)
	db := setupDB(t)
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11)
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
	db := setupDB(t)
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11)
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
	db := setupDB(t)
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11)
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
	db := setupDB(t)
	defer func() { _ = db.Close() }()
	user := auth.NewMockUser(1, 11)
	dataStore := database.NewDataStore(db)
	itemStore := dataStore.Items()

	var result []itemIdRow
	db = itemStore.VisibleGrandChildrenOfID(user, 190).Select("items.ID").Scan(&result)
	assert.NoError(db.Error())

	expected := []itemIdRow{{ID: 19000}, {ID: 19001}, {ID: 19002}}
	assert.Equal(expected, result)
}
