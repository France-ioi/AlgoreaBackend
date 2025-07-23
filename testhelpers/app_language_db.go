//go:build !prod && !unit

package testhelpers

import (
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func (ctx *TestContext) databaseCountRows(table string, datamap map[string]string) int {
	query := ctx.application.Database.Table(table)
	for key, value := range datamap {
		columnName := database.QuoteName(key)
		switch {
		case value == nullValue:
			query = query.Where(columnName + " IS NULL")
		case value == tableValueFalse:
			query = query.Where(columnName + " = 0")
		case value == tableValueTrue:
			query = query.Where(columnName + " = 1")
		default:
			var processedValue interface{} = value
			if value[0] == referencePrefix {
				processedValue = ctx.getIDOfReference(value)
			}
			query = query.Where(columnName+" = ?", processedValue)
		}
	}

	var resultCount int
	mustNotBeError(query.Count(&resultCount).Error())

	return resultCount
}
