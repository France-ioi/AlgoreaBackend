//go:build !prod

package testhelpers

func (ctx *TestContext) databaseCountRows(table string, datamap map[string]string) int {
	query, values := ctx.buildDatabaseCountRowQuery(table, datamap)

	return ctx.queryScalar(query, values)
}

// buildDatabaseCountRowQuery builds a query to count the rows in a table that match the map.
func (ctx *TestContext) buildDatabaseCountRowQuery(table string, datamap map[string]string) (query string, values []interface{}) {
	var conditions string
	for key, value := range datamap {
		if conditions != "" {
			conditions += " AND "
		}

		switch {
		case value == nullValue:
			conditions += "`" + key + "` IS NULL "
		case value == tableValueFalse:
			conditions += "`" + key + "` = 0 "
		case value == tableValueTrue:
			conditions += "`" + key + "` = 1 "
		default:
			conditions += "`" + key + "`" + " = ? "
			if value[0] == referencePrefix {
				values = append(values, ctx.getIDOfReference(value))
			} else {
				values = append(values, value)
			}
		}
	}

	table = "`" + table + "`"
	query = "SELECT COUNT(*) as count FROM " + table + " WHERE " + conditions

	return query, values
}

// queryScalar returns a single value from the database as the result of the query.
func (ctx *TestContext) queryScalar(query string, values []interface{}) int {
	var resultCount int
	err := ctx.db.
		QueryRow(query, values...).
		Scan(&resultCount)
	mustNotBeError(err)

	return resultCount
}
