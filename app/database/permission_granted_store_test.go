package database

import (
	"errors"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPermissionGrantedStore_After_MustBeInTransaction(t *testing.T) {
	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()

	assert.PanicsWithValue(t, ErrNoTransaction, func() {
		_ = NewDataStore(db).PermissionsGranted().After()
	})

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestPermissionGrantedStore_After_HandlesErrorOfComputeAllAccess(t *testing.T) {
	expectedError := errors.New("some error")

	db, dbMock := NewDBMock()
	defer func() { _ = db.Close() }()
	dbMock.ExpectBegin()
	dbMock.ExpectPrepare("^DROP TEMPORARY TABLE IF EXISTS").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).PermissionsGranted().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestPermissionGrantedStore_ViewIndexByKind(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		viewIndexes = map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"}
		viewKinds = map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5}
	})
	defer monkey.UnpatchAll()
	defer clearAllPermissionEnums()

	assert.Equal(t, 5, permissionGrantedStore.ViewIndexByKind("solution"))
	assert.Panics(t, func() { permissionGrantedStore.ViewIndexByKind("unknown") })
}

func TestPermissionGrantedStore_ViewKindByIndex(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		viewIndexes = map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"}
		viewKinds = map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5}
	})
	defer monkey.UnpatchAll()
	defer clearAllPermissionEnums()

	assert.Equal(t, "solution", permissionGrantedStore.ViewKindByIndex(5))
	assert.Equal(t, "content_with_descendants", permissionGrantedStore.ViewKindByIndex(4))
	assert.Panics(t, func() { permissionGrantedStore.ViewKindByIndex(10) })
}

func TestPermissionGrantedStore_ViewKindByIndex_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockPermissionEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	defer clearAllPermissionEnums()

	assert.Equal(t, "solution", permissionGrantedStore.ViewKindByIndex(5))
}

func TestPermissionGrantedStore_GrantViewIndexByKind(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		grantViewKinds = map[string]int{"none": 1, "content": 2, "content_with_descendants": 3, "solution": 4, "transfer": 5}
		grantViewIndexes = map[int]string{1: "none", 2: "content", 3: "content_with_descendants", 4: "solution", 5: "transfer"}
	})
	defer monkey.UnpatchAll()
	defer clearAllPermissionEnums()

	assert.Panics(t, func() { permissionGrantedStore.GrantViewIndexByKind("unknown") })
	assert.Equal(t, 5, permissionGrantedStore.GrantViewIndexByKind("transfer"))
	assert.Equal(t, 4, permissionGrantedStore.GrantViewIndexByKind("solution"))
}

func TestPermissionGrantedStore_GrantViewIndexByKind_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockPermissionEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	defer clearAllPermissionEnums()

	assert.Equal(t, 5, permissionGrantedStore.GrantViewIndexByKind("transfer"))
}

func TestPermissionGrantedStore_EditIndexByKind(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		editIndexes = map[int]string{1: "none", 2: "children", 3: "all", 4: "transfer"}
		editKinds = map[string]int{"none": 1, "children": 2, "all": 3, "transfer": 4}
	})
	defer monkey.UnpatchAll()
	defer clearAllPermissionEnums()

	assert.Equal(t, 4, permissionGrantedStore.EditIndexByKind("transfer"))
	assert.Equal(t, 3, permissionGrantedStore.EditIndexByKind("all"))
	assert.Panics(t, func() { permissionGrantedStore.EditIndexByKind("unknown") })
}

func TestPermissionGrantedStore_EditIndexByKind_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockPermissionEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllPermissionEnums()
	defer clearAllPermissionEnums()

	assert.Equal(t, 4, permissionGrantedStore.EditIndexByKind("transfer"))
}

func mockPermissionEnumQueries(sqlMock sqlmock.Sqlmock) {
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_view").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','info','content','content_with_descendants','solution'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_grant_view").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','content','content_with_descendants','solution','transfer'"))
	sqlMock.ExpectQuery("^"+regexp.QuoteMeta(
		"SELECT SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6) FROM `information_schema`.`COLUMNS`  "+
			"WHERE (TABLE_SCHEMA = DATABASE()) AND (TABLE_NAME = ?) AND (COLUMN_NAME = ?) LIMIT 1")+"$").
		WithArgs("permissions_granted", "can_edit").
		WillReturnRows(sqlMock.NewRows([]string{"value"}).
			AddRow("'none','children','all','transfer'"))
}

func clearAllPermissionEnums() {
	viewKinds = nil
	viewIndexes = nil
	grantViewKinds = nil
	grantViewIndexes = nil
	editKinds = nil
	editIndexes = nil
}