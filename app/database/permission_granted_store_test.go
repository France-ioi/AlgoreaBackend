package database

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
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
	dbMock.ExpectPrepare("^INSERT INTO permissions_propagate").WillReturnError(expectedError)
	dbMock.ExpectRollback()

	assert.Equal(t, expectedError, db.inTransaction(func(trDB *DB) error {
		return NewDataStore(trDB).PermissionsGranted().After()
	}))

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestPermissionGrantedStore_ViewIndexByName(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		fakeDBEnums("permissions_granted.can_view",
			map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5},
			map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"})
	})
	defer monkey.UnpatchAll()
	defer clearAllDBEnums()

	assert.Equal(t, 5, permissionGrantedStore.ViewIndexByName("solution"))
	assert.Panics(t, func() { permissionGrantedStore.ViewIndexByName("unknown") })
}

func TestPermissionGrantedStore_ViewNameByIndex(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		fakeDBEnums("permissions_granted.can_view",
			map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5},
			map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"})
	})
	defer monkey.UnpatchAll()
	defer clearAllDBEnums()

	assert.Equal(t, "solution", permissionGrantedStore.ViewNameByIndex(5))
	assert.Equal(t, "content_with_descendants", permissionGrantedStore.ViewNameByIndex(4))
	assert.Panics(t, func() { permissionGrantedStore.ViewNameByIndex(10) })
}

func TestPermissionGrantedStore_ViewNameByIndex_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockDBEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	defer clearAllDBEnums()

	assert.Equal(t, "solution", permissionGrantedStore.ViewNameByIndex(5))
}

func TestPermissionGrantedStore_GrantViewEnum(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		fakeDBEnums("permissions_granted.can_grant_view",
			map[string]int{"none": 1, "content": 2, "content_with_descendants": 3, "solution": 4, "solution_with_grant": 5},
			map[int]string{1: "none", 2: "content", 3: "content_with_descendants", 4: "solution", 5: "solution_with_grant"})
	})
	defer monkey.UnpatchAll()
	defer clearAllDBEnums()

	assert.Equal(t, 5, permissionGrantedStore.GrantViewIndexByName("solution_with_grant"))
	assert.Equal(t, 3, permissionGrantedStore.GrantViewIndexByName("content_with_descendants"))
	assert.Panics(t, func() { permissionGrantedStore.GrantViewIndexByName("unknown") })
	assert.Equal(t, 4, permissionGrantedStore.GrantViewIndexByName("solution"))
	clearAllDBEnums()
	assert.Equal(t, "solution_with_grant", permissionGrantedStore.GrantViewNameByIndex(5))
	assert.Equal(t, "content_with_descendants", permissionGrantedStore.GrantViewNameByIndex(3))
	assert.Panics(t, func() { permissionGrantedStore.GrantViewNameByIndex(0) })
	assert.Equal(t, "solution", permissionGrantedStore.GrantViewNameByIndex(4))
}

func TestPermissionGrantedStore_GrantViewIndexByName_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockDBEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	defer clearAllDBEnums()

	assert.Equal(t, 6, permissionGrantedStore.GrantViewIndexByName("solution_with_grant"))
}

func TestPermissionGrantedStore_EditEnum(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	var oldLock *monkey.PatchGuard
	oldLock = monkey.PatchInstanceMethod(reflect.TypeOf(&sync.RWMutex{}), "Lock", func(mutex *sync.RWMutex) {
		oldLock.Unpatch()
		mutex.Lock()
		oldLock.Restore()
		fakeDBEnums("permissions_granted.can_view",
			map[string]int{"none": 1, "info": 2, "content": 3, "content_with_descendants": 4, "solution": 5},
			map[int]string{1: "none", 2: "info", 3: "content", 4: "content_with_descendants", 5: "solution"})
		fakeDBEnums("permissions_granted.can_edit",
			map[string]int{"none": 1, "children": 2, "all": 3, "all_with_grant": 4},
			map[int]string{1: "none", 2: "children", 3: "all", 4: "all_with_grant"})
	})
	defer monkey.UnpatchAll()
	defer clearAllDBEnums()

	assert.Equal(t, 4, permissionGrantedStore.EditIndexByName("all_with_grant"))
	assert.Equal(t, 3, permissionGrantedStore.EditIndexByName("all"))
	assert.Panics(t, func() { permissionGrantedStore.EditIndexByName("unknown") })
	clearAllDBEnums()
	assert.Equal(t, "all_with_grant", permissionGrantedStore.EditNameByIndex(4))
	assert.Equal(t, "all", permissionGrantedStore.EditNameByIndex(3))
	assert.Panics(t, func() { permissionGrantedStore.EditNameByIndex(0) })
}

func TestPermissionGrantedStore_EditIndexByName_Load(t *testing.T) {
	db, sqlMock := NewDBMock()
	defer func() { _ = db.Close() }()

	mockDBEnumQueries(sqlMock)
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	defer clearAllDBEnums()

	assert.Equal(t, 4, permissionGrantedStore.EditIndexByName("all_with_grant"))
}

func TestPermissionGrantedStore_WatchEnum(t *testing.T) {
	db, _ := NewDBMock()
	defer func() { _ = db.Close() }()
	permissionsGrantedStore := NewDataStore(db).PermissionsGranted()

	clearAllDBEnums()
	fakeDBEnums("permissions_granted.can_watch",
		map[string]int{"none": 1, "result": 2, "answer": 3, "answer_with_grant": 4},
		map[int]string{1: "none", 2: "result", 3: "answer", 4: "answer_with_grant"})
	defer clearAllDBEnums()

	assert.Equal(t, 4, permissionsGrantedStore.WatchIndexByName("answer_with_grant"))
	assert.Equal(t, 3, permissionsGrantedStore.WatchIndexByName("answer"))
	assert.Equal(t, 2, permissionsGrantedStore.WatchIndexByName("result"))
	assert.Equal(t, 1, permissionsGrantedStore.WatchIndexByName("none"))
	assert.Panics(t, func() { permissionsGrantedStore.WatchIndexByName("unknown") })
	clearAllDBEnums()
	fakeDBEnums("permissions_granted.can_watch",
		map[string]int{"none": 1, "result": 2, "answer": 3, "answer_with_grant": 4},
		map[int]string{1: "none", 2: "result", 3: "answer", 4: "answer_with_grant"})
	assert.Equal(t, "answer_with_grant", permissionsGrantedStore.WatchNameByIndex(4))
	assert.Equal(t, "answer", permissionsGrantedStore.WatchNameByIndex(3))
	assert.Equal(t, "result", permissionsGrantedStore.WatchNameByIndex(2))
	assert.Equal(t, "none", permissionsGrantedStore.WatchNameByIndex(1))
	assert.Panics(t, func() { permissionsGrantedStore.WatchNameByIndex(10) })
}
