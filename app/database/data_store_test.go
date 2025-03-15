package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func TestDataStore_StoreConstructorsSetTablesCorrectly(t *testing.T) {
	tests := []struct {
		name      string
		function  func(store *DataStore) *DB
		wantTable string
	}{
		{"Answers", func(store *DataStore) *DB { return store.Answers().Where("") }, "`answers`"},
		{"Attempts", func(store *DataStore) *DB { return store.Attempts().Where("") }, "`attempts`"},
		{"Gradings", func(store *DataStore) *DB { return store.Gradings().Where("") }, "`gradings`"},
		{"Groups", func(store *DataStore) *DB { return store.Groups().Where("") }, "`groups`"},
		{"GroupAncestors", func(store *DataStore) *DB { return store.GroupAncestors().Where("") }, "`groups_ancestors`"},
		{"ActiveGroupAncestors", func(store *DataStore) *DB { return store.ActiveGroupAncestors().Where("") }, "`groups_ancestors_active`"},
		{"GroupGroups", func(store *DataStore) *DB { return store.GroupGroups().Where("") }, "`groups_groups`"},
		{"ActiveGroupGroups", func(store *DataStore) *DB { return store.ActiveGroupGroups().Where("") }, "`groups_groups_active`"},
		{"GroupMembershipChanges", func(store *DataStore) *DB { return store.GroupMembershipChanges().Where("") }, "`group_membership_changes`"},
		{
			"GroupItemAdditionalTimes", func(store *DataStore) *DB { return store.GroupItemAdditionalTimes().Where("") },
			"`group_item_additional_times`",
		},
		{"GroupManagers", func(store *DataStore) *DB { return store.GroupManagers().Where("") }, "`group_managers`"},
		{"GroupPendingRequests", func(store *DataStore) *DB { return store.GroupPendingRequests().Where("") }, "`group_pending_requests`"},
		{"Permissions", func(store *DataStore) *DB { return store.Permissions().Where("") }, "permissions_generated AS permissions"},
		{"PermissionsGranted", func(store *DataStore) *DB { return store.PermissionsGranted().Where("") }, "`permissions_granted`"},
		{"Items", func(store *DataStore) *DB { return store.Items().Where("") }, "`items`"},
		{"ItemAncestors", func(store *DataStore) *DB { return store.ItemAncestors().Where("") }, "`items_ancestors`"},
		{"ItemItems", func(store *DataStore) *DB { return store.ItemItems().Where("") }, "`items_items`"},
		{"ItemStrings", func(store *DataStore) *DB { return store.ItemStrings().Where("") }, "`items_strings`"},
		{"ItemDependencies", func(store *DataStore) *DB { return store.ItemDependencies().Where("") }, "`item_dependencies`"},
		{"Languages", func(store *DataStore) *DB { return store.Languages().Where("") }, "`languages`"},
		{"Platforms", func(store *DataStore) *DB { return store.Platforms().Where("") }, "`platforms`"},
		{"Results", func(store *DataStore) *DB { return store.Results().Where("") }, "`results`"},
		{"Sessions", func(store *DataStore) *DB { return store.Sessions().Where("") }, "`sessions`"},
		{"AccessTokens", func(store *DataStore) *DB { return store.AccessTokens().Where("") }, "`access_tokens`"},
		{"Threads", func(store *DataStore) *DB { return store.Threads().Where("") }, "`threads`"},
		{"Users", func(store *DataStore) *DB { return store.Users().Where("") }, "`users`"},
		{"UserBatches", func(store *DataStore) *DB { return store.UserBatches().Where("") }, "`user_batches_v2`"},
		{"UserBatchPrefixes", func(store *DataStore) *DB { return store.UserBatchPrefixes().Where("") }, "`user_batch_prefixes`"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testoutput.SuppressIfPasses(t)

			db, mock := NewDBMock()
			defer func() { _ = db.Close() }()
			mock.ExpectQuery("SELECT \\* FROM " + tt.wantTable).
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
		{"Attempts", func(store *DataStore) interface{} { return store.Attempts() }, &AttemptStore{}},
		{"Gradings", func(store *DataStore) interface{} { return store.Gradings() }, &GradingStore{}},
		{"Groups", func(store *DataStore) interface{} { return store.Groups() }, &GroupStore{}},
		{"GroupAncestors", func(store *DataStore) interface{} { return store.GroupAncestors() }, &GroupAncestorStore{}},
		{"ActiveGroupAncestors", func(store *DataStore) interface{} { return store.ActiveGroupAncestors() }, &GroupAncestorStore{}},
		{"GroupGroups", func(store *DataStore) interface{} { return store.GroupGroups() }, &GroupGroupStore{}},
		{"ActiveGroupGroups", func(store *DataStore) interface{} { return store.ActiveGroupGroups() }, &GroupGroupStore{}},
		{"GroupMembershipChanges", func(store *DataStore) interface{} { return store.GroupMembershipChanges() }, &GroupMembershipChangeStore{}},
		{
			"GroupItemAdditionalTimes", func(store *DataStore) interface{} { return store.GroupItemAdditionalTimes() },
			&GroupItemAdditionalTimeStore{},
		},
		{"GroupManagers", func(store *DataStore) interface{} { return store.GroupManagers() }, &GroupManagerStore{}},
		{"GroupPendingRequests", func(store *DataStore) interface{} { return store.GroupPendingRequests() }, &GroupPendingRequestStore{}},
		{"Permissions", func(store *DataStore) interface{} { return store.Permissions() }, &PermissionGeneratedStore{}},
		{"PermissionsGranted", func(store *DataStore) interface{} { return store.PermissionsGranted() }, &PermissionGrantedStore{}},
		{"Items", func(store *DataStore) interface{} { return store.Items() }, &ItemStore{}},
		{"ItemAncestors", func(store *DataStore) interface{} { return store.ItemAncestors() }, &ItemAncestorStore{}},
		{"ItemItems", func(store *DataStore) interface{} { return store.ItemItems() }, &ItemItemStore{}},
		{"ItemStrings", func(store *DataStore) interface{} { return store.ItemStrings() }, &ItemStringStore{}},
		{"ItemDependencies", func(store *DataStore) interface{} { return store.ItemDependencies() }, &ItemDependencyStore{}},
		{"Languages", func(store *DataStore) interface{} { return store.Languages() }, &LanguageStore{}},
		{"Platforms", func(store *DataStore) interface{} { return store.Platforms() }, &PlatformStore{}},
		{"Results", func(store *DataStore) interface{} { return store.Results() }, &ResultStore{}},
		{"Sessions", func(store *DataStore) interface{} { return store.Sessions() }, &SessionStore{}},
		{"AccessTokens", func(store *DataStore) interface{} { return store.AccessTokens() }, &AccessTokenStore{}},
		{"Threads", func(store *DataStore) interface{} { return store.Threads() }, &ThreadStore{}},
		{"Users", func(store *DataStore) interface{} { return store.Users() }, &UserStore{}},
		{"UserBatches", func(store *DataStore) interface{} { return store.UserBatches() }, &UserBatchStore{}},
		{"UserBatchPrefixes", func(store *DataStore) interface{} { return store.UserBatchPrefixes() }, &UserBatchPrefixStore{}},
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
	testoutput.SuppressIfPasses(t)

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
	testoutput.SuppressIfPasses(t)

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
	testoutput.SuppressIfPasses(t)

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

func TestDataStore_InTransaction_ContextAndTxOptions(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	var callsCount int

	type ctxKey string

	txOptions := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	patch := patchGormBeginTxWithVerifier(t, &callsCount, txOptions, map[interface{}]interface{}{ctxKey("key"): "value"})
	defer patch.Unpatch()

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	db.ctx = context.WithValue(context.Background(), ctxKey("key"), "value")

	mock.ExpectBegin()
	mock.ExpectCommit()

	store := NewDataStoreWithTable(db, "myTable")
	gotError := store.InTransaction(func(s *DataStore) error {
		return nil
	}, txOptions)

	assert.Nil(t, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, callsCount)
}

func TestDataStore_InTransaction_ForcesTransactionRetryingForTestingPurposes(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	store := NewDataStoreWithContext(ContextWithTransactionRetrying(context.Background()), db)
	gotError := store.InTransaction(func(s *DataStore) error {
		var result []interface{}
		return s.Raw("SELECT 1").Scan(&result).Error()
	})

	assert.NoError(t, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_InTransaction_ForcesTransactionRetryingForTestingPurposes_Hooks(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	var called []string
	SetOnStartOfTransactionToBeRetriedForcefullyHook(func() {
		called = append(called, "start")
	})
	defer SetOnStartOfTransactionToBeRetriedForcefullyHook(func() {})
	SetOnForcefulRetryOfTransactionHook(func() {
		called = append(called, "retry")
	})
	defer SetOnForcefulRetryOfTransactionHook(func() {})

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	store := NewDataStoreWithContext(ContextWithTransactionRetrying(context.Background()), db)
	gotError := store.InTransaction(func(s *DataStore) error {
		var result []interface{}
		called = append(called, "1")
		return s.Raw("SELECT 1").Scan(&result).Error()
	})

	assert.NoError(t, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, []string{"start", "1", "retry", "1"}, called)
}

func TestDataStore_WithForeignKeyChecksDisabled_DBErrorOnStartingTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("^SET").WillReturnError(expectedError)
	mock.ExpectRollback()

	store := NewDataStore(db)
	gotError := store.WithForeignKeyChecksDisabled(func(*DataStore) error {
		assert.Fail(t, "should not be called")
		return nil
	})

	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithForeignKeyChecksDisabled_WithTxOptions(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	var callsCount int
	txOptions := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	patch := patchGormBeginTxWithVerifier(t, &callsCount, txOptions, nil)
	defer patch.Unpatch()

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectExec("^SET").WithArgs().WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectExec("^SET").WithArgs().WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectCommit()

	store := NewDataStore(db)
	gotError := store.WithForeignKeyChecksDisabled(func(*DataStore) error {
		return nil
	}, txOptions)

	assert.Nil(t, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, callsCount)
}

func TestDataStore_WithForeignKeyChecksDisabled_DBErrorOnCommittingTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectQuery("^SELECT 1").WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectCommit().WillReturnError(expectedError)

	store := NewDataStore(db)
	gotError := store.WithForeignKeyChecksDisabled(func(s *DataStore) error {
		var result []interface{}
		return s.Raw("SELECT 1").Scan(&result).Error()
	})

	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithForeignKeyChecksDisabled_DBErrorInsideTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectRollback()

	store := NewDataStore(db)
	gotError := store.WithForeignKeyChecksDisabled(func(*DataStore) error {
		return expectedError
	})

	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithForeignKeyChecksDisabled_DBErrorWithoutTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	expectedError := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectExec("^SET").WillReturnResult(sqlmock.NewResult(-1, -1))
	mock.ExpectRollback()

	store := NewDataStore(db)
	gotError := store.InTransaction(func(innerStore *DataStore) error {
		return innerStore.WithForeignKeyChecksDisabled(func(*DataStore) error {
			return expectedError
		})
	})

	assert.Equal(t, expectedError, gotError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithNamedLock(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	lockName := "some lock name"
	timeout := 1234 * time.Millisecond
	expectedTimeout := int(timeout.Round(time.Second).Seconds())
	assertNamedLockMethod(t, lockName, expectedTimeout, "tableName",
		func(store *DataStore) func(func(store *DataStore) error) error {
			return func(txFunc func(store *DataStore) error) error {
				return store.WithNamedLock(lockName, timeout, txFunc)
			}
		})
}

func assertNamedLockMethod(t *testing.T, expectedLockName string, expectedTimeout int, expectedTableName string,
	funcToTestGenerator func(store *DataStore) func(func(store *DataStore) error) error,
) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	dbMock.ExpectQuery("^"+regexp.QuoteMeta("SELECT GET_LOCK(?, ?)")+"$").
		WithArgs(expectedLockName, expectedTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"GET_LOCK(?, ?)"}).AddRow(int64(1)))
	dbMock.ExpectQuery("SELECT 1 AS id").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	dbMock.ExpectQuery("^" + regexp.QuoteMeta("SELECT RELEASE_LOCK(?)") + "$").
		WithArgs(expectedLockName).WillReturnRows(sqlmock.NewRows([]string{"RELEASE_LOCK(?)"}).AddRow(int64(1)))

	store := NewDataStoreWithTable(db, "tableName")
	err := funcToTestGenerator(store)(func(s *DataStore) error {
		assert.Equal(t, expectedTableName, s.tableName)
		assert.NotEqual(t, store, s)
		assert.Equal(t, store.db.CommonDB(), s.db.CommonDB())
		var result []interface{}
		return db.Raw("SELECT 1 AS id").Scan(&result).Error()
	})
	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_RetryOnDuplicatePrimaryKeyError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	for i := 1; i < keyTriesCount; i++ {
		dbMock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'users.PRIMARY'"})
	}
	dbMock.ExpectExec(retryOnDuplicatePrimaryKeyErrorExpectedQueryRegexp).WithArgs(keyTriesCount).
		WillReturnResult(sqlmock.NewResult(keyTriesCount, 1))

	retryCount := 0
	err := NewDataStore(db).RetryOnDuplicatePrimaryKeyError("users", func(store *DataStore) error {
		retryCount++
		return db.Exec("INSERT INTO users (id) VALUES (?)", retryCount).Error()
	})

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_RetryOnDuplicateKeyError(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	queryRegexp := "^" + regexp.QuoteMeta("INSERT INTO users (login) VALUES (?)") + "$"
	for i := 1; i < keyTriesCount; i++ {
		dbMock.ExpectExec(queryRegexp).WithArgs(i).
			WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry '" + strconv.Itoa(i) + "' for key 'users.login'"})
	}
	dbMock.ExpectExec(queryRegexp).WithArgs(keyTriesCount).
		WillReturnResult(sqlmock.NewResult(keyTriesCount, 1))

	retryCount := 0
	err := NewDataStore(db).RetryOnDuplicateKeyError("users", "login", "login", func(store *DataStore) error {
		retryCount++
		return db.Exec("INSERT INTO users (login) VALUES (?)", retryCount).Error()
	})

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_InsertMap(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRow := map[string]interface{}{"id": int64(1), "sField": "some value", "sNullField": nil}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?)")).
		WithArgs(int64(1), "some value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, NewDataStoreWithTable(db, "myTable").InsertMap(dataRow))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_InsertOrUpdateMap(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRow := map[string]interface{}{"id": int64(1), "sField": "some value", "sNullField": nil}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?) "+
		"ON DUPLICATE KEY UPDATE `sField` = VALUES(`sField`), `sNullField` = VALUES(`sNullField`)")).
		WithArgs(int64(1), "some value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, NewDataStoreWithTable(db, "myTable").
		InsertOrUpdateMap(dataRow, []string{"sField", "sNullField"}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_InsertOrUpdateMaps(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataRows := []map[string]interface{}{
		{"id": int64(1), "sField": "some value", "sNullField": "value"},
		{"id": int64(2), "sField": "another value", "sNullField": nil},
	}

	expectedError := errors.New("some error")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `myTable` (`id`, `sField`, `sNullField`) VALUES (?, ?, ?), (?, ?, ?) "+
		"ON DUPLICATE KEY UPDATE `sField` = VALUES(`sField`), `sNullField` = VALUES(`sNullField`)")).
		WithArgs(int64(1), "some value", "value", int64(2), "another value", nil).
		WillReturnError(expectedError)

	assert.Equal(t, expectedError, NewDataStoreWithTable(db, "myTable").
		InsertOrUpdateMaps(dataRows, []string{"sField", "sNullField"}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithSharedWriteLock(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `myTable` FOR SHARE").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	dataStore := NewDataStoreWithTable(db, "myTable")
	err := dataStore.inTransaction(func(db *DB) error {
		newDataStore := NewDataStore(db).WithSharedWriteLock()
		assert.NotEqual(t, newDataStore, dataStore)
		assert.NoError(t, newDataStore.Error())
		var result []interface{}
		assert.NoError(t, newDataStore.Scan(&result).Error())
		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithSharedWriteLock_PanicsWhenNotInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() { NewDataStore(db).WithSharedWriteLock() })
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithCustomWriteLock(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(
		"SELECT `t1`.\\* FROM `t1` JOIN `t2` JOIN `t3` JOIN `t4` FOR SHARE OF `t1`, `t2` FOR UPDATE OF `t3`, `t4`").
		WillReturnRows(mock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectCommit()

	dataStore := NewDataStoreWithTable(db, "t1").Joins("JOIN `t2`").Joins("JOIN `t3`").Joins("JOIN `t4`")
	err := dataStore.inTransaction(func(db *DB) error {
		newDataStore := NewDataStore(db).WithCustomWriteLocks(golang.NewSet("t1", "t2"), golang.NewSet("t3", "t4"))
		assert.NotEqual(t, newDataStore, dataStore)
		assert.NoError(t, newDataStore.Error())
		var result []interface{}
		assert.NoError(t, newDataStore.Scan(&result).Error())
		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_WithCustomWriteLock_PanicsWhenNotInTransaction(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		NewDataStore(db).WithCustomWriteLocks(
			golang.NewSet[string](), golang.NewSet[string]())
	})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_PropagationsSchedules_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		NewDataStore(db).SchedulePermissionsPropagation()
	})
	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		NewDataStore(db).ScheduleResultsPropagation()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestProhibitResultsPropagation(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	dataStore := NewDataStore(db)
	assert.False(t, dataStore.IsResultsPropagationProhibited())

	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	ProhibitResultsPropagation(db)
	assert.True(t, dataStore.IsResultsPropagationProhibited())
	assert.NoError(t, dataStore.InTransaction(func(dataStore *DataStore) error {
		dataStore.ScheduleResultsPropagation()
		return nil
	}))
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDataStore_MergeContext(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()

	dataStoreWithEmptyContext := NewDataStore(db)
	newContext := dataStoreWithEmptyContext.MergeContext(context.Background())
	assert.Equal(t, propagationsBitField{}, newContext.Value(prohibitedPropagationsContextKey))

	expectedBitField := propagationsBitField{
		Permissions: false,
		Results:     true,
	}
	dataStoreWithProhibitedResultsPropagation := NewDataStoreWithContext(
		context.WithValue(context.Background(), prohibitedPropagationsContextKey, expectedBitField), db)
	newContext = dataStoreWithProhibitedResultsPropagation.MergeContext(context.Background())
	assert.Equal(t, expectedBitField, newContext.Value(prohibitedPropagationsContextKey))
}

func TestDataStore_IsInTransaction_ReturnsTrue(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		assert.True(t, store.isInTransaction())
		return nil
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_IsInTransaction_ReturnsFalse(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.False(t, NewDataStore(db).IsInTransaction())
	assert.NoError(t, mock.ExpectationsWereMet())
}

type gormDialectDBAccessor struct {
	_ unsafe.Pointer
	v *struct {
		db gorm.SQLCommon
		gorm.DefaultForeignKeyNamer
	}
}

type testContextKey string

func TestNewDataStoreWithContext_WithSQLDBWrapper(t *testing.T) {
	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	ctx := context.WithValue(context.Background(), testContextKey("key"), "value")
	dataStore := NewDataStoreWithContext(ctx, db)

	assert.Equal(t, db.ctes, dataStore.ctes)
	assert.Equal(t, db.logConfig, dataStore.logConfig)
	assert.Equal(t, ctx, dataStore.ctx)

	dbWrapper := dataStore.DB.db.CommonDB().(*sqlDBWrapper)
	assert.Equal(t, ctx, dbWrapper.ctx)
	assert.Equal(t, db.logConfig, dbWrapper.logConfig)
	assert.Equal(t, db.db.CommonDB().(*sqlDBWrapper).sqlDB, dbWrapper.sqlDB)
	dialect := dataStore.DB.db.Dialect()
	assert.Equal(t, dbWrapper, (*gormDialectDBAccessor)(unsafe.Pointer(&dialect)).v.db)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestNewDataStoreWithContext_WithSQLTxWrapper(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := db.inTransaction(func(db *DB) error {
		ctx := context.WithValue(context.Background(), testContextKey("key"), "value")
		dataStore := NewDataStoreWithContext(ctx, db)

		assert.Equal(t, db.ctes, dataStore.ctes)
		assert.Equal(t, db.logConfig, dataStore.logConfig)
		assert.Equal(t, ctx, dataStore.ctx)

		txWrapper := dataStore.DB.db.CommonDB().(*sqlTxWrapper)
		assert.Equal(t, db.logConfig, txWrapper.logConfig)
		assert.Equal(t, ctx, txWrapper.ctx)
		assert.Equal(t, db.db.CommonDB().(*sqlTxWrapper).sqlTx, txWrapper.sqlTx)
		dialect := dataStore.DB.db.Dialect()
		assert.Equal(t, txWrapper, (*gormDialectDBAccessor)(unsafe.Pointer(&dialect)).v.db)

		return nil
	})
	require.NoError(t, err)

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestDataStore_EnsureTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	assert.NoError(t, NewDataStore(db).EnsureTransaction(func(store *DataStore) error {
		assert.True(t, store.isInTransaction())
		return nil
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_EnsureTransaction_InsideTransaction(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		return store.EnsureTransaction(func(store *DataStore) error {
			assert.True(t, store.isInTransaction())
			return nil
		})
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDataStore_SetPropagationsModeToSync(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	db, mock := NewDBMock()
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectExec("^" + regexp.QuoteMeta("SET @synchronous_propagations_connection_id = CONNECTION_ID()") + "$").
		WillReturnResult(sqlmock.NewResult(-1, 0))
	mock.ExpectCommit()

	require.Nil(t, db.ctx.Value(propagationsAreSyncContextKey))

	assert.NoError(t, NewDataStore(db).InTransaction(func(store *DataStore) error {
		require.NoError(t, store.SetPropagationsModeToSync())
		assert.Equal(t, store.DB.ctx.Value(propagationsAreSyncContextKey), true)
		assert.Equal(t, store.DB.db.CommonDB().(*sqlTxWrapper).ctx.Value(propagationsAreSyncContextKey), true)
		return nil
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}
