package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"bou.ke/monkey"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func (ctx *TestContext) DBHasTable(tableName string, data *gherkin.DataTable) error { // nolint
	db := ctx.db()

	if len(data.Rows) > 1 {
		head := data.Rows[0].Cells
		fields := make([]string, 0, len(head))
		marks := make([]string, 0, len(head))
		for _, cell := range head {
			fields = append(fields, cell.Value)
			marks = append(marks, "?")
		}

		marksString := "(" + strings.Join(marks, ", ") + ")"
		finalMarksString := marksString
		if len(data.Rows) > 2 {
			finalMarksString = strings.Repeat(marksString+", ", len(data.Rows)-2) + finalMarksString
		}
		query := "INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES " + finalMarksString // nolint: gosec
		vals := make([]interface{}, 0, (len(data.Rows)-1)*len(head))
		for i := 1; i < len(data.Rows); i++ {
			for _, cell := range data.Rows[i].Cells {
				var err error
				if cell.Value, err = ctx.preprocessString(cell.Value); err != nil {
					return err
				}
				vals = append(vals, dbDataTableValue(cell.Value))
			}
		}
		if ctx.inScenario {
			_, err := db.Exec(query, vals...)
			if err != nil {
				return err
			}
		} else {
			ctx.featureQueries = append(ctx.featureQueries, dbquery{query, vals})
		}
	}

	if ctx.dbTableData[tableName] == nil {
		ctx.dbTableData[tableName] = data
	} else if len(data.Rows) > 1 {
		ctx.dbTableData[tableName] = combineGherkinTables(ctx.dbTableData[tableName], data)
	}

	return nil
}

func (ctx *TestContext) TableShouldBeEmpty(tableName string) error { // nolint
	db := ctx.db()
	sqlRows, err := db.Query(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", tableName)) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = sqlRows.Close() }()
	if sqlRows.Next() {
		return fmt.Errorf("the table %q should be empty, but it is not", tableName)
	}

	return nil
}

func (ctx *TestContext) TableShouldBe(tableName string, data *gherkin.DataTable) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, nil, false, data)
}

func (ctx *TestContext) TableShouldStayUnchanged(tableName string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &gherkin.DataTable{Rows: []*gherkin.TableRow{
			{Cells: []*gherkin.TableCell{{Value: "1"}}}},
		}
	}
	return ctx.tableAtIDShouldBe(tableName, nil, false, data)
}

func (ctx *TestContext) TableShouldStayUnchangedButTheRowWithID(tableName string, ids string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &gherkin.DataTable{Rows: []*gherkin.TableRow{}}
	}
	return ctx.tableAtIDShouldBe(tableName, parseMultipleIDString(ids), true, data)
}

func (ctx *TestContext) TableAtIDShouldBe(tableName string, ids string, data *gherkin.DataTable) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, parseMultipleIDString(ids), false, data)
}

func (ctx *TestContext) TableShouldNotContainID(tableName string, ids string) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, parseMultipleIDString(ids), false,
		&gherkin.DataTable{Rows: []*gherkin.TableRow{{Cells: []*gherkin.TableCell{{Value: "ID"}}}}})
}

func (ctx *TestContext) TableHasUniqueKey(tableName, indexName, columns string) error { // nolint
	db, err := gorm.Open("mysql", ctx.db())
	if err != nil {
		return err
	}

	if db.Dialect().HasIndex(tableName, indexName) {
		return nil
	}

	if err := db.Table(tableName).AddUniqueIndex(indexName, strings.Split(columns, ",")...).Error; err != nil {
		return err
	}
	ctx.addedDBIndices = append(ctx.addedDBIndices, &addedDBIndex{Table: tableName, Index: indexName})
	return nil
}

func combineGherkinTables(table1, table2 *gherkin.DataTable) *gherkin.DataTable {
	table1FieldMap := map[string]int{}
	combinedFieldMap := map[string]bool{}
	columnNumber := len(table1.Rows[0].Cells)
	combinedColumnNames := make([]string, 0, columnNumber+len(table2.Rows[0].Cells))
	for index, cell := range table1.Rows[0].Cells {
		table1FieldMap[cell.Value] = index
		combinedFieldMap[cell.Value] = true
		combinedColumnNames = append(combinedColumnNames, cell.Value)
	}
	table2FieldMap := map[string]int{}
	for index, cell := range table2.Rows[0].Cells {
		table2FieldMap[cell.Value] = index
		// only add a column if it hasn't been met in table1
		if !combinedFieldMap[cell.Value] {
			combinedFieldMap[cell.Value] = true
			columnNumber++
			combinedColumnNames = append(combinedColumnNames, cell.Value)
		}
	}

	combinedTable := &gherkin.DataTable{}
	combinedTable.Rows = make([]*gherkin.TableRow, 0, len(table1.Rows)+len(table2.Rows)-1)

	header := &gherkin.TableRow{Cells: make([]*gherkin.TableCell, 0, columnNumber)}
	for _, columnName := range combinedColumnNames {
		header.Cells = append(header.Cells, &gherkin.TableCell{Value: columnName})
	}
	combinedTable.Rows = append(combinedTable.Rows, header)

	copyCellsIntoCombinedTable(table1, combinedColumnNames, table1FieldMap, combinedTable)
	copyCellsIntoCombinedTable(table2, combinedColumnNames, table2FieldMap, combinedTable)
	return combinedTable
}

func copyCellsIntoCombinedTable(sourceTable *gherkin.DataTable, combinedColumnNames []string,
	sourceTableFieldMap map[string]int, combinedTable *gherkin.DataTable) {
	for rowNum := 1; rowNum < len(sourceTable.Rows); rowNum++ {
		newRow := &gherkin.TableRow{Cells: make([]*gherkin.TableCell, 0, len(combinedColumnNames))}
		for _, column := range combinedColumnNames {
			var newCell *gherkin.TableCell
			if sourceColumnNumber, ok := sourceTableFieldMap[column]; ok {
				newCell = sourceTable.Rows[rowNum].Cells[sourceColumnNumber]
			}
			newRow.Cells = append(newRow.Cells, newCell)
		}
		combinedTable.Rows = append(combinedTable.Rows, newRow)
	}
}

func parseMultipleIDString(idsString string) []int64 {
	split := strings.Split(idsString, ",")
	ids := make([]int64, 0, len(split))
	for _, idString := range split {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func (ctx *TestContext) tableAtIDShouldBe(tableName string, ids []int64, excludeIDs bool, data *gherkin.DataTable) error { // nolint
	// For that, we build a SQL request with only the attribute we are interested about (those
	// for the test data table) and we convert them to string (in SQL) to compare to table value.
	// Expect 'null' string in the table to check for nullness

	db := ctx.db()

	var selects []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		selects = append(selects, cell.Value)
	}

	idsMap := make(map[string]bool, len(ids))
	for _, id := range ids {
		idsMap[strconv.FormatInt(id, 10)] = true
	}
	idsStrings := make([]string, 0, len(ids))
	for idString := range idsMap {
		idsStrings = append(idsStrings, idString)
	}
	idsString := strings.Join(idsStrings, ",")
	// define 'where' condition if needed
	where := ""
	if len(ids) > 0 {
		if excludeIDs {
			where = fmt.Sprintf(" WHERE ID NOT IN (%s) ", idsString) // #nosec
		} else {
			where = fmt.Sprintf(" WHERE ID IN (%s) ", idsString) // #nosec
		}
	}

	selectsJoined := strings.Join(selects, ", ")

	// exec sql
	query := fmt.Sprintf("SELECT %s FROM `%s` %s ORDER BY %s", selectsJoined, tableName, where, selectsJoined) // nolint: gosec
	sqlRows, err := db.Query(query)
	defer func() { _ = sqlRows.Close() }()
	if err != nil {
		return err
	}
	dataCols := data.Rows[0].Cells
	idColumnIndex := -1
	for index, cell := range dataCols {
		if cell.Value == "ID" {
			idColumnIndex = index
			break
		}
	}

	iDataRow := 1
	sqlCols, _ := sqlRows.Columns() // nolint: gosec
	for sqlRows.Next() {
		for excludeIDs && iDataRow < len(data.Rows) && idsMap[data.Rows[iDataRow].Cells[idColumnIndex].Value] {
			iDataRow++
		}
		if iDataRow >= len(data.Rows) {
			return fmt.Errorf("there are more rows in the SQL results than expected. expected: %d", len(data.Rows)-1)
		}
		// Create a slice of string to represent each attribute value,
		// and a second slice to contain pointers to each item.
		rowValues := make([]*string, len(sqlCols))
		rowValPtr := make([]interface{}, len(sqlCols))
		for i := range rowValues {
			rowValPtr[i] = &rowValues[i]
		}
		// Scan the result into the column pointers...
		if err := sqlRows.Scan(rowValPtr...); err != nil {
			return err
		}

		nullValue := tableValueNull
		pNullValue := &nullValue
		// checking that all columns of the test data table match the SQL row
		for iCol, dataCell := range data.Rows[iDataRow].Cells {
			if dataCell == nil {
				continue
			}
			colName := dataCols[iCol].Value
			dataValue, err := ctx.preprocessString(dataCell.Value)
			if err != nil {
				return err
			}
			sqlValue := rowValPtr[iCol].(**string)

			if *sqlValue == nil {
				sqlValue = &pNullValue
			}

			if (dataValue == tableValueTrue && **sqlValue == "1") || (dataValue == tableValueFalse && **sqlValue == "0") {
				continue
			}

			if dataValue != **sqlValue {
				return fmt.Errorf("not matching expected value at row %d, col %s, expected '%s', got: '%v'", iDataRow, colName, dataValue, **sqlValue)
			}
		}

		iDataRow++
	}

	for excludeIDs && iDataRow < len(data.Rows) && idsMap[data.Rows[iDataRow].Cells[idColumnIndex].Value] {
		iDataRow++
	}

	// check that no row in the test data table has not been uncheck (if less rows in SQL result)
	if iDataRow < len(data.Rows) {
		return fmt.Errorf("there are less rows in the SQL results than expected")
	}
	return nil
}

var nowRegexp = regexp.MustCompile(`(?i)\bNOW\s*\(\s*\)`)

func (ctx *TestContext) DbTimeNow(timeStrRaw string) error { // nolint
	timeStr := fmt.Sprintf("%q", timeStrRaw)

	// patch database.DB's methods
	standardDBMethods := [...]string{
		"Where", "Or", "Select", "Having",
	}
	standardDBGuards := make(map[string]*monkey.PatchGuard, len(standardDBMethods))
	for _, methodName := range standardDBMethods {
		methodName := methodName
		standardDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query interface{}, args ...interface{}) *database.DB {
				standardDBGuards[methodName].Unpatch()
				defer standardDBGuards[methodName].Restore()
				if queryStr, ok := query.(string); ok {
					query = nowRegexp.ReplaceAllString(queryStr, timeStr)
				}
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := make([]reflect.Value, 0, len(args))
				reflArgs = append(reflArgs, reflect.ValueOf(query))
				for _, arg := range args {
					arg := arg
					reflArgs = append(reflArgs, reflect.ValueOf(arg))
				}

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
	}

	stringAndArgsDBMethods := [...]string{
		"Joins", "Raw", "Exec",
	}
	stringAndArgsDBGuards := make(map[string]*monkey.PatchGuard, len(stringAndArgsDBMethods))
	for _, methodName := range stringAndArgsDBMethods {
		methodName := methodName
		stringAndArgsDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query string, args ...interface{}) *database.DB {
				stringAndArgsDBGuards[methodName].Unpatch()
				defer stringAndArgsDBGuards[methodName].Restore()
				query = nowRegexp.ReplaceAllString(query, timeStr)
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := make([]reflect.Value, 0, len(args))
				reflArgs = append(reflArgs, reflect.ValueOf(query))
				for _, arg := range args {
					arg := arg
					reflArgs = append(reflArgs, reflect.ValueOf(arg))
				}

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
	}

	stringDBMethods := [...]string{
		"Table", "Group",
	}
	stringDBGuards := make(map[string]*monkey.PatchGuard, len(stringDBMethods))
	for _, methodName := range stringDBMethods {
		methodName := methodName
		stringDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query string) *database.DB {
				stringDBGuards[methodName].Unpatch()
				defer stringDBGuards[methodName].Restore()
				query = nowRegexp.ReplaceAllString(query, timeStr)
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := []reflect.Value{reflect.ValueOf(query)}

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
	}

	interfaceDBMethods := [...]string{
		"Union", "UnionAll",
	}
	interfaceDBGuards := make(map[string]*monkey.PatchGuard, len(interfaceDBMethods))
	for _, methodName := range interfaceDBMethods {
		methodName := methodName
		interfaceDBGuards[methodName] = monkey.PatchInstanceMethod(
			reflect.TypeOf(&database.DB{}), methodName,
			func(db *database.DB, query interface{}) *database.DB {
				interfaceDBGuards[methodName].Unpatch()
				defer interfaceDBGuards[methodName].Restore()
				if queryStr, ok := query.(string); ok {
					query = nowRegexp.ReplaceAllString(queryStr, timeStr)
				}
				reflMethod := reflect.ValueOf(db).MethodByName(methodName)
				reflArgs := []reflect.Value{reflect.ValueOf(query)}

				return reflMethod.Call(reflArgs)[0].Interface().(*database.DB)
			})
	}

	var orderGuard *monkey.PatchGuard
	orderGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Order",
		func(db *database.DB, value interface{}, reorder ...bool) *database.DB {
			orderGuard.Unpatch()
			defer orderGuard.Restore()
			if valueStr, ok := value.(string); ok {
				value = nowRegexp.ReplaceAllString(valueStr, timeStr)
			}
			return db.Order(value, reorder...)
		})

	var pluckGuard *monkey.PatchGuard
	pluckGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Pluck",
		func(db *database.DB, column string, values interface{}) *database.DB {
			pluckGuard.Unpatch()
			defer pluckGuard.Restore()
			column = nowRegexp.ReplaceAllString(column, timeStr)
			return db.Pluck(column, values)
		})

	var takeGuard *monkey.PatchGuard
	takeGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Take",
		func(db *database.DB, out interface{}, where ...interface{}) *database.DB {
			takeGuard.Unpatch()
			defer takeGuard.Restore()
			if len(where) > 0 {
				if whereStr, ok := where[0].(string); ok {
					where[0] = nowRegexp.ReplaceAllString(whereStr, timeStr)
				}
			}
			return db.Take(out, where...)
		})

	var deleteGuard *monkey.PatchGuard
	deleteGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&database.DB{}), "Delete",
		func(db *database.DB, where ...interface{}) *database.DB {
			deleteGuard.Unpatch()
			defer deleteGuard.Restore()
			if len(where) > 0 {
				if whereStr, ok := where[0].(string); ok {
					where[0] = nowRegexp.ReplaceAllString(whereStr, timeStr)
				}
			}
			return db.Delete(where...)
		})

	database.MockNow(timeStrRaw)

	// Patch Gorm's methods
	var execGuard, rawGuard, prepareContextGuard, queryContextGuard *monkey.PatchGuard
	execGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&gorm.DB{}), "Exec",
		func(db *gorm.DB, query string, args ...interface{}) *gorm.DB {
			execGuard.Unpatch()
			defer execGuard.Restore()
			query = nowRegexp.ReplaceAllString(query, timeStr)
			return db.Exec(query, args...)
		})
	rawGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&gorm.DB{}), "Raw",
		func(db *gorm.DB, query string, args ...interface{}) *gorm.DB {
			rawGuard.Unpatch()
			defer rawGuard.Restore()
			query = nowRegexp.ReplaceAllString(query, timeStr)
			return db.Raw(query, args...)
		})

	// db methods
	prepareContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&sql.DB{}), "PrepareContext",
		func(db *sql.DB, c context.Context, query string) (*sql.Stmt, error) {
			prepareContextGuard.Unpatch()
			defer prepareContextGuard.Restore()
			query = nowRegexp.ReplaceAllString(query, timeStr)
			return db.PrepareContext(c, query)
		})
	queryContextGuard = monkey.PatchInstanceMethod(
		reflect.TypeOf(&sql.DB{}), "QueryContext",
		func(db *sql.DB, c context.Context, query string, args ...interface{}) (*sql.Rows, error) {
			queryContextGuard.Unpatch()
			defer queryContextGuard.Restore()
			query = nowRegexp.ReplaceAllString(query, timeStr)
			return db.QueryContext(c, query, args...)
		})

	return nil
}

const tableValueFalse = "false"
const tableValueTrue = "true"
const tableValueNull = "null"

// dbDataTableValue converts a string value that we can find the db seeding table to a valid type for the db
// e.g., the string "null" means the SQL `NULL`
func dbDataTableValue(input string) interface{} {
	switch input {
	case tableValueFalse:
		return false
	case tableValueTrue:
		return true
	case tableValueNull:
		return nil
	default:
		return input
	}
}
