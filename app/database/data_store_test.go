package database

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestDataStore_StoreConstructorsSetTablesCorrectly(t *testing.T) {
	tests := []struct {
		name      string
		function  func(store *DataStore) *DB
		wantTable string
	}{
		{"Answers", func(store *DataStore) *DB { return store.Answers().Where("") }, "`answers`"},
		{"Groups", func(store *DataStore) *DB { return store.Groups().Where("") }, "`groups`"},
		{"GroupAncestors", func(store *DataStore) *DB { return store.GroupAncestors().Where("") }, "`groups_ancestors`"},
		{"ActiveGroupAncestors", func(store *DataStore) *DB { return store.ActiveGroupAncestors().Where("") }, "`groups_ancestors_active`"},
		{"GroupAttempts", func(store *DataStore) *DB { return store.GroupAttempts().Where("") }, "`groups_attempts`"},
		{"GroupGroups", func(store *DataStore) *DB { return store.GroupGroups().Where("") }, "`groups_groups`"},
		{"ActiveGroupGroups", func(store *DataStore) *DB { return store.ActiveGroupGroups().Where("") }, "`groups_groups_active`"},
		{"GroupMembershipChanges", func(store *DataStore) *DB { return store.GroupMembershipChanges().Where("") }, "`group_membership_changes`"},
		{"GroupContestItems", func(store *DataStore) *DB { return store.GroupContestItems().Where("") }, "`groups_contest_items`"},
		{"GroupManagers", func(store *DataStore) *DB { return store.GroupManagers().Where("") }, "`group_managers`"},
		{"GroupPendingRequests", func(store *DataStore) *DB { return store.GroupPendingRequests().Where("") }, "`group_pending_requests`"},
		{"Permissions", func(store *DataStore) *DB { return store.Permissions().Where("") }, "permissions_generated AS permissions"},
		{"PermissionsGranted", func(store *DataStore) *DB { return store.PermissionsGranted().Where("") }, "`permissions_granted`"},
		{"Items", func(store *DataStore) *DB { return store.Items().Where("") }, "`items`"},
		{"ItemAncestors", func(store *DataStore) *DB { return store.ItemAncestors().Where("") }, "`items_ancestors`"},
		{"ItemItems", func(store *DataStore) *DB { return store.ItemItems().Where("") }, "`items_items`"},
		{"ItemStrings", func(store *DataStore) *DB { return store.ItemStrings().Where("") }, "`items_strings`"},
		{"ItemUnlockingRules", func(store *DataStore) *DB { return store.ItemUnlockingRules().Where("") }, "`item_unlocking_rules`"},
		{"Languages", func(store *DataStore) *DB { return store.Languages().Where("") }, "`languages`"},
		{"LoginStates", func(store *DataStore) *DB { return store.LoginStates().Where("") }, "`login_states`"},
		{"Platforms", func(store *DataStore) *DB { return store.Platforms().Where("") }, "`platforms`"},
		{"RefreshTokens", func(store *DataStore) *DB { return store.RefreshTokens().Where("") }, "`refresh_tokens`"},
		{"Sessions", func(store *DataStore) *DB { return store.Sessions().Where("") }, "`sessions`"},
		{"Users", func(store *DataStore) *DB { return store.Users().Where("") }, "`users`"},
		{"UserItems", func(store *DataStore) *DB { return store.UserItems().Where("") }, "`users_items`"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()
			mock.ExpectQuery("SELECT \\* FROM " + tt.wantTable). // nolint:gosec
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
		{"Answers", func(store *DataStore) interface{} { return store.Answers() }, &AnswerStore{}},
		{"Groups", func(store *DataStore) interface{} { return store.Groups() }, &GroupStore{}},
		{"GroupAncestors", func(store *DataStore) interface{} { return store.GroupAncestors() }, &GroupAncestorStore{}},
		{"ActiveGroupAncestors", func(store *DataStore) interface{} { return store.ActiveGroupAncestors() }, &GroupAncestorStore{}},
		{"GroupAttempts", func(store *DataStore) interface{} { return store.GroupAttempts() }, &GroupAttemptStore{}},
		{"GroupGroups", func(store *DataStore) interface{} { return store.GroupGroups() }, &GroupGroupStore{}},
		{"ActiveGroupGroups", func(store *DataStore) interface{} { return store.ActiveGroupGroups() }, &GroupGroupStore{}},
		{"GroupMembershipChanges", func(store *DataStore) interface{} { return store.GroupMembershipChanges() }, &GroupMembershipChangeStore{}},
		{"GroupContestItems", func(store *DataStore) interface{} { return store.GroupContestItems() }, &GroupContestItemStore{}},
		{"GroupManagers", func(store *DataStore) interface{} { return store.GroupManagers() }, &GroupManagerStore{}},
		{"GroupPendingRequests", func(store *DataStore) interface{} { return store.GroupPendingRequests() }, &GroupPendingRequestStore{}},
		{"Permissions", func(store *DataStore) interface{} { return store.Permissions() }, &PermissionGeneratedStore{}},
		{"PermissionsGranted", func(store *DataStore) interface{} { return store.PermissionsGranted() }, &PermissionGrantedStore{}},
		{"Items", func(store *DataStore) interface{} { return store.Items() }, &ItemStore{}},
		{"ItemAncestors", func(store *DataStore) interface{} { return store.ItemAncestors() }, &ItemAncestorStore{}},
		{"ItemItems", func(store *DataStore) interface{} { return store.ItemItems() }, &ItemItemStore{}},
		{"ItemStrings", func(store *DataStore) interface{} { return store.ItemStrings() }, &ItemStringStore{}},
		{"ItemUnlockingRules", func(store *DataStore) interface{} { return store.ItemUnlockingRules() }, &ItemUnlockingRuleStore{}},
		{"Languages", func(store *DataStore) interface{} { return store.Languages() }, &LanguageStore{}},
		{"LoginStates", func(store *DataStore) interface{} { return store.LoginStates() }, &LoginStateStore{}},
		{"Platforms", func(store *DataStore) interface{} { return store.Platforms() }, &PlatformStore{}},
		{"RefreshTokens", func(store *DataStore) interface{} { return store.RefreshTokens() }, &RefreshTokenStore{}},
		{"Sessions", func(store *DataStore) interface{} { return store.Sessions() }, &SessionStore{}},
		{"Users", func(store *DataStore) interface{} { return store.Users() }, &UserStore{}},
		{"UserItems", func(store *DataStore) interface{} { return store.UserItems() }, &UserItemStore{}},
	}
	for _, tt := range tests {
		tt := tt
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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `tableName` WHERE (tableName.id = ?)")).
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"id"}))

	var result []interface{}
	err := NewDataStoreWithTable(db, "tableName").ByID(id).Scan(&result).Error()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_ByID_ForAbstractDataStore(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, "method ByID() called for abstract DataStore", func() {
		NewDataStore(db).ByID(123)
	})
}

func TestDataStore_InTransaction_NoErrors(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 AS id").
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()

	type resultStruct struct {
		ID int64
	}

	store := NewDataStoreWithTable(db, "myTable")
	var result []resultStruct
	err := store.InTransaction(func(s *DataStore) error {
		assert.Equal(t, store.tableName, s.tableName)
		assert.NotEqual(t, store, s)
		assert.NotEqual(t, store.db, s.db)

		return db.Raw("SELECT 1 AS id").Scan(&result).Error()
	})

	assert.NoError(t, err)
	assert.Equal(t, []resultStruct{{1}}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_InTransaction_DBError(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnError(expectedError)
	mock.ExpectRollback()

	store := NewDataStoreWithTable(db, "myTable")
	gotError := store.InTransaction(func(s *DataStore) error {
		assert.Equal(t, store.tableName, s.tableName)
		assert.NotEqual(t, store, s)
		assert.NotEqual(t, store.db, s.db)

		var result []interface{}
		return db.Raw("SELECT 1").Scan(&result).Error()
	})

	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithNamedLock(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	lockName := "some lock name"
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(lockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
	dbMock.ExpectQuery("SELECT 1 AS id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	dbMock.ExpectExec("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs(lockName).WillReturnResult(sqlmock.NewResult(-1, -1))

	store := NewDataStoreWithTable(db, "tableName")
	err := store.WithNamedLock(lockName, timeout, func(s *DataStore) error {
		assert.Equal(t, store.tableName, s.tableName)
		assert.Equal(t, store, s)
		assert.Equal(t, store.db, s.db)
		var result []interface{}
		return db.Raw("SELECT 1 AS id").Scan(&result).Error()
	})
	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_RetryOnDuplicatePrimaryKeyError(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	for i := 1; i < keyTriesCount; i++ {
		dbMock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'PRIMARY'"})
	}
	dbMock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(keyTriesCount).
		WillReturnResult(sqlmock.NewResult(keyTriesCount, 1))

	retryCount := 0
	err := NewDataStore(db).RetryOnDuplicatePrimaryKeyError(func(store *DataStore) error {
		retryCount++
		return db.Exec("INSERT INTO users (id) VALUES (?)", retryCount).Error()
	})

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_RetryOnDuplicateKeyError(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	queryRegexp := "^" + regexp.QuoteMeta("INSERT INTO users (login) VALUES (?)") + "$"
	for i := 1; i < keyTriesCount; i++ {
		dbMock.ExpectExec(queryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'login'"})
	}
	dbMock.ExpectExec(queryRegexp).WithArgs(keyTriesCount).
		WillReturnResult(sqlmock.NewResult(keyTriesCount, 1))

	retryCount := 0
	err := NewDataStore(db).RetryOnDuplicateKeyError("login", "login", func(store *DataStore) error {
		retryCount++
		return db.Exec("INSERT INTO users (login) VALUES (?)", retryCount).Error()
	})

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_InsertMap(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRow := map[string]interface{}{"id": int64(1), "sField": "some value", "sNullField": nil}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, NULL)")).
		WithArgs(int64(1), "some value").
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, NewDataStoreWithTable(db, "myTable").InsertMap(dataRow))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_InsertOrUpdateMap(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRow := map[string]interface{}{"id": int64(1), "sField": "some value", "sNullField": nil}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, NULL) "+
		"ON DUPLICATE KEY UPDATE `sField` = VALUES(`sField`), `sNullField` = VALUES(`sNullField`)")).
		WithArgs(int64(1), "some value").
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, NewDataStoreWithTable(db, "myTable").
		InsertOrUpdateMap(dataRow, []string{"sField", "sNullField"}))
	assert.NoError(t, mock.ExpectationsWereMet())
}
