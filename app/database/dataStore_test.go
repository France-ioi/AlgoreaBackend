package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataStore_StoreConstructorsSetTablesCorrectly(t *testing.T) {
	tests := []struct {
		name      string
		function  func(store *DataStore) *DB
		wantTable string
	}{
		{"Groups", func(store *DataStore) *DB { return store.Groups().Where("") }, "groups"},
		{"GroupAncestors", func(store *DataStore) *DB { return store.GroupAncestors().Where("") }, "groups_ancestors"},
		{"GroupAttempts", func(store *DataStore) *DB { return store.GroupAttempts().Where("") }, "groups_attempts"},
		{"GroupGroups", func(store *DataStore) *DB { return store.GroupGroups().Where("") }, "groups_groups"},
		{"GroupItems", func(store *DataStore) *DB { return store.GroupItems().Where("") }, "groups_items"},
		{"Items", func(store *DataStore) *DB { return store.Items().Where("") }, "items"},
		{"ItemAncestors", func(store *DataStore) *DB { return store.ItemAncestors().Where("") }, "items_ancestors"},
		{"ItemItems", func(store *DataStore) *DB { return store.ItemItems().Where("") }, "items_items"},
		{"ItemStrings", func(store *DataStore) *DB { return store.ItemStrings().Where("") }, "items_strings"},
		{"Users", func(store *DataStore) *DB { return store.Users().Where("") }, "users"},
		{"UserAnswers", func(store *DataStore) *DB { return store.UserAnswers().Where("") }, "users_answers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()
			mock.ExpectQuery("SELECT \\* FROM `" + tt.wantTable + "`").
				WithArgs().WillReturnRows(mock.NewRows([]string{"id"}))

			var result []interface{}
			assert.NoError(t, tt.function(NewDataStore(db)).Scan(&result).Error())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDataStore_StoreConstructorsReturnObjectsOfRightTypes(t *testing.T) {
	tests := []struct {
		name     string
		function func(store *DataStore) interface{}
		wantType interface{}
	}{
		{"Groups", func(store *DataStore) interface{} { return store.Groups() }, &GroupStore{}},
		{"GroupAncestors", func(store *DataStore) interface{} { return store.GroupAncestors() }, &GroupAncestorStore{}},
		{"GroupAttempts", func(store *DataStore) interface{} { return store.GroupAttempts() }, &GroupAttemptStore{}},
		{"GroupGroups", func(store *DataStore) interface{} { return store.GroupGroups() }, &GroupGroupStore{}},
		{"GroupItems", func(store *DataStore) interface{} { return store.GroupItems() }, &GroupItemStore{}},
		{"Items", func(store *DataStore) interface{} { return store.Items() }, &ItemStore{}},
		{"ItemAncestors", func(store *DataStore) interface{} { return store.ItemAncestors() }, &ItemAncestorStore{}},
		{"ItemItems", func(store *DataStore) interface{} { return store.ItemItems() }, &ItemItemStore{}},
		{"ItemStrings", func(store *DataStore) interface{} { return store.ItemStrings() }, &ItemStringStore{}},
		{"Users", func(store *DataStore) interface{} { return store.Users() }, &UserStore{}},
		{"UserAnswers", func(store *DataStore) interface{} { return store.UserAnswers() }, &UserAnswerStore{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := NewDBMock()
			defer func() { _ = db.Close() }()

			store := tt.function(NewDataStore(db))
			assert.IsType(t, tt.wantType, store)
		})
	}
}

func TestDataStore_ByID(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	const id = 123
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `tableName` WHERE (tableName.ID = ?)")).
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"ID"}))

	var result []interface{}
	err := NewDataStoreWithTable(db, "tableName").ByID(id).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
