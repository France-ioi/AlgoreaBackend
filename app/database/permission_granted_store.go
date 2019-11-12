package database

import (
	"fmt"
	"strings"
	"sync"
)

// PermissionGrantedStore implements database operations on `permissions_granted`
type PermissionGrantedStore struct {
	*DataStore
}

var (
	enumsMutex       sync.RWMutex
	viewNames        map[string]int
	viewIndexes      map[int]string
	grantViewNames   map[string]int
	grantViewIndexes map[int]string
	editNames        map[string]int
	editIndexes      map[int]string
)

// After is a "listener" that calls PermissionGrantedStore::computeAllAccess()
func (s *PermissionGrantedStore) After() (err error) {
	s.mustBeInTransaction()
	defer recoverPanics(&err)

	s.computeAllAccess()
	return nil
}

func (s *PermissionGrantedStore) removePartialAccess(groupID, itemID int64) {
	mustNotBeError(s.Where("item_id = ?", itemID).Where("group_id = ?", groupID).
		Where("giver_group_id = -1").Where("can_view = 'content'").
		Where("is_owner = 0").
		UpdateColumn("can_view", "none").Error())
}

// PermissionIndexByKindAndName returns the index of the given permission in the enum
func (s *PermissionGrantedStore) PermissionIndexByKindAndName(kind, name string) int {
	permissionMap := map[string]*map[string]int{
		"view":       &viewNames,
		"grant_view": &grantViewNames,
		"edit":       &editNames,
	}[kind]
	getterFunc := func() int { return requireIndexByName(*permissionMap, name, "can_"+kind) }
	enumsMutex.RLock()
	if len(*permissionMap) != 0 {
		defer enumsMutex.RUnlock()
		return getterFunc()
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(*permissionMap) != 0 {
		return getterFunc()
	}

	s.loadViewKinds()
	return getterFunc()
}

// ViewIndexByName returns the index of the given view kind in the 'can_view' enum
func (s *PermissionGrantedStore) ViewIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("view", name)
}

// PermissionNameByKindAndIndex returns the permission name of the given kind with the given index from the enum
func (s *PermissionGrantedStore) PermissionNameByKindAndIndex(kind string, index int) string {
	permissionMap := map[string]*map[int]string{
		"view":       &viewIndexes,
		"grant_view": &grantViewIndexes,
		"edit":       &editIndexes,
	}[kind]
	getterFunc := func() string { return requireNameByIndex(*permissionMap, index, "can_"+kind) }
	enumsMutex.RLock()
	if len(*permissionMap) != 0 {
		defer enumsMutex.RUnlock()
		return getterFunc()
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(*permissionMap) != 0 {
		return getterFunc()
	}

	s.loadViewKinds()
	return getterFunc()
}

// ViewNameByIndex returns the view permission name with the given index from the 'can_view' enum
func (s *PermissionGrantedStore) ViewNameByIndex(index int) string {
	return s.PermissionNameByKindAndIndex("view", index)
}

func (s *PermissionGrantedStore) loadViewKinds() {
	viewNames, viewIndexes = s.loadKindsIntoMaps("permissions_granted", "can_view")
	grantViewNames, grantViewIndexes = s.loadKindsIntoMaps("permissions_granted", "can_grant_view")
	editNames, editIndexes = s.loadKindsIntoMaps("permissions_granted", "can_edit")
}

func (s *PermissionGrantedStore) loadKindsIntoMaps(tableName, columnName string) (kindsMap map[string]int, indexesMap map[int]string) {
	var valuesString string
	mustNotBeError(NewDataStore(newDB(s.db.New())).Table("information_schema.COLUMNS").
		Set("gorm:query_option", "").
		Where("TABLE_SCHEMA = DATABASE()").
		Where("TABLE_NAME = ?", tableName).
		Where("COLUMN_NAME = ?", columnName).
		PluckFirst("SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6)", &valuesString).Error())
	values := strings.Split(valuesString, ",")
	kindsMap = make(map[string]int, len(values))
	indexesMap = make(map[int]string, len(values))
	for index, value := range values {
		kind := strings.Trim(value, "'")
		realIndex := index + 1 // 0 is reserved for an empty value
		kindsMap[kind] = realIndex
		indexesMap[realIndex] = kind
	}
	return kindsMap, indexesMap
}

// GrantViewIndexByName returns the index of the given "grant view" permission name in the 'can_grant_view' enum
func (s *PermissionGrantedStore) GrantViewIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("grant_view", name)
}

// EditIndexByName returns the index of the given "edit" permission name in the 'can_edit' enum
func (s *PermissionGrantedStore) EditIndexByName(name string) int {
	return s.PermissionIndexByKindAndName("edit", name)
}

func requireIndexByName(m map[string]int, name, kind string) int {
	if index, ok := m[name]; ok {
		return index
	}
	panic(fmt.Errorf("unknown permission %s for %s", name, kind))
}

func requireNameByIndex(m map[int]string, index int, kind string) string {
	if name, ok := m[index]; ok {
		return name
	}
	panic(fmt.Errorf("wrong index %d for %s", index, kind))
}
