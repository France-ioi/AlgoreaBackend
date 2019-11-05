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
	viewKinds        map[string]int
	viewIndexes      map[int]string
	grantViewKinds   map[string]int
	grantViewIndexes map[int]string
	editKinds        map[string]int
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

// ViewIndexByKind returns the index of the given view kind in the 'can_view' enum
func (s *PermissionGrantedStore) ViewIndexByKind(kind string) int {
	enumsMutex.RLock()
	if len(viewKinds) != 0 {
		defer enumsMutex.RUnlock()
		return requireIndexByKind(viewKinds, kind, "can_view")
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(viewKinds) != 0 {
		return requireIndexByKind(viewKinds, kind, "can_view")
	}

	s.loadViewKinds()
	return requireIndexByKind(viewKinds, kind, "can_view")
}

// ViewKindByIndex returns the view kind with the given index from the 'can_view' enum
func (s *PermissionGrantedStore) ViewKindByIndex(index int) string {
	enumsMutex.RLock()
	if len(viewIndexes) != 0 {
		defer enumsMutex.RUnlock()
		return requireKindByIndex(viewIndexes, index, "can_view")
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(viewIndexes) != 0 {
		return requireKindByIndex(viewIndexes, index, "can_view")
	}

	s.loadViewKinds()
	return requireKindByIndex(viewIndexes, index, "can_view")
}

func (s *PermissionGrantedStore) loadViewKinds() {
	viewKinds, viewIndexes = s.loadKindsIntoMaps("permissions_granted", "can_view")
	grantViewKinds, grantViewIndexes = s.loadKindsIntoMaps("permissions_granted", "can_grant_view")
	editKinds, editIndexes = s.loadKindsIntoMaps("permissions_granted", "can_edit")
}

func (s *PermissionGrantedStore) loadKindsIntoMaps(tableName, columnName string) (kindsMap map[string]int, indexesMap map[int]string) {
	var valuesString string
	mustNotBeError(s.Table("information_schema.COLUMNS").
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

// GrantViewIndexByKind returns the index of the given "grant view" kind in the 'can_grant_view' enum
func (s *PermissionGrantedStore) GrantViewIndexByKind(kind string) int {
	enumsMutex.RLock()
	if len(grantViewKinds) != 0 {
		defer enumsMutex.RUnlock()
		return requireIndexByKind(grantViewKinds, kind, "can_grant_view")
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(grantViewKinds) != 0 {
		return requireIndexByKind(grantViewKinds, kind, "can_grant_view")
	}

	s.loadViewKinds()
	return requireIndexByKind(grantViewKinds, kind, "can_grant_view")
}

// EditIndexByKind returns the index of the given "edit" kind in the 'can_edit' enum
func (s *PermissionGrantedStore) EditIndexByKind(kind string) int {
	enumsMutex.RLock()
	if len(editKinds) != 0 {
		defer enumsMutex.RUnlock()
		return requireIndexByKind(editKinds, kind, "can_edit")
	}
	enumsMutex.RUnlock()

	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	if len(editKinds) != 0 {
		return requireIndexByKind(editKinds, kind, "can_edit")
	}

	s.loadViewKinds()
	return requireIndexByKind(editKinds, kind, "can_edit")
}

func requireIndexByKind(m map[string]int, kind, name string) int {
	if index, ok := m[kind]; ok {
		return index
	}
	panic(fmt.Errorf("unknown kind %s for %s", kind, name))
}

func requireKindByIndex(m map[int]string, index int, name string) string {
	if kind, ok := m[index]; ok {
		return kind
	}
	panic(fmt.Errorf("wrong index %d for %s", index, name))
}
