package database

import (
	"fmt"
	"strings"
	"sync"
)

var (
	enumsMutex  sync.RWMutex
	enumColumns = []string{
		"permissions_granted.can_view",
		"permissions_granted.can_grant_view",
		"permissions_granted.can_watch",
		"permissions_granted.can_edit",
		"group_managers.can_manage",
	}
	enumName2Number     map[string]int
	enumValueName2Index []map[string]int
	enumValueIndex2Name []map[int]string
)

func loadDBEnum(db *DB, fullColumnName string) {
	enumName2Number = make(map[string]int, len(enumColumns))
	for index := range enumColumns {
		enumName2Number[enumColumns[index]] = index
	}

	parsedColumn := strings.SplitN(fullColumnName, ".", 2)
	tableName := parsedColumn[0]
	columnName := parsedColumn[1]

	var valuesString string
	mustNotBeError(NewDataStore(newDB(db.db.New())).Table("information_schema.COLUMNS").
		Set("gorm:query_option", "").
		Where("TABLE_SCHEMA = DATABASE()").
		Where("TABLE_NAME = ?", tableName).
		Where("COLUMN_NAME = ?", columnName).
		PluckFirst("SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE)-6)", &valuesString).Error())
	values := strings.Split(valuesString, ",")
	namesMap := make(map[string]int, len(values))
	indexesMap := make(map[int]string, len(values))
	for index, value := range values {
		kind := strings.Trim(value, "'")
		realIndex := index + 1 // 0 is reserved for an empty value
		namesMap[kind] = realIndex
		indexesMap[realIndex] = kind
	}
	enumNumber := enumName2Number[fullColumnName]
	enumValueName2Index[enumNumber] = namesMap
	enumValueIndex2Name[enumNumber] = indexesMap
}

func getFromDBEnumUnderLock(db *DB, getterFunc func() interface{}) interface{} {
	// Lock for reading to check if the enums have been already loaded
	enumsMutex.RLock()
	if len(enumValueName2Index) != 0 { // the enums have been loaded, so return the value
		defer enumsMutex.RUnlock()
		return getterFunc()
	}
	enumsMutex.RUnlock()

	// Lock for writing to load the enums from the DB
	enumsMutex.Lock()
	defer enumsMutex.Unlock()
	// Check if the enums have been loaded while we were waiting for the lock
	if len(enumValueName2Index) != 0 {
		return getterFunc() // the enums have been loaded, so return the value
	}

	var success bool
	defer func() {
		if !success {
			enumValueIndex2Name = nil
			enumValueName2Index = nil
		}
	}()
	loadAllEnums(db)
	success = true

	return getterFunc()
}

func loadAllEnums(db *DB) {
	enumValueName2Index = make([]map[string]int, len(enumColumns))
	enumValueIndex2Name = make([]map[int]string, len(enumColumns))
	for _, fullColumnName := range enumColumns {
		loadDBEnum(db, fullColumnName)
	}
}

func requireDBEnumIndexByName(enumName, name string) int {
	if index, ok := enumValueName2Index[enumName2Number[enumName]][name]; ok {
		return index
	}
	panic(fmt.Errorf("unknown enum value %s for %s", name, enumName))
}

func requireDBEnumNameByIndex(enumName string, index int) string {
	if name, ok := enumValueIndex2Name[enumName2Number[enumName]][index]; ok {
		return name
	}
	panic(fmt.Errorf("wrong enum index %d for %s", index, enumName))
}
