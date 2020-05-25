// +build !prod

package testhelpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/messages-go/v10"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func (ctx *TestContext) DBHasTable(tableName string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	db := ctx.db()

	if len(data.Rows) > 1 {
		head := data.Rows[0].Cells
		fields := make([]string, 0, len(head))
		marks := make([]string, 0, len(head))
		for _, cell := range head {
			fields = append(fields, database.QuoteName(cell.Value))
			marks = append(marks, "?")
		}

		marksString := "(" + strings.Join(marks, ", ") + ")"
		finalMarksString := marksString
		if len(data.Rows) > 2 {
			finalMarksString = strings.Repeat(marksString+", ", len(data.Rows)-2) + finalMarksString
		}
		query := "INSERT INTO " + database.QuoteName(tableName) + // nolint: gosec
			" (" + strings.Join(fields, ", ") + ") VALUES " + finalMarksString
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
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0")
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			_, err = tx.Exec(query, vals...)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			err = tx.Commit()
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
		ctx.dbTableData[tableName] = combinePickleTables(ctx.dbTableData[tableName], data)
	}

	return nil
}

func (ctx *TestContext) DBHasUsers(data *messages.PickleStepArgument_PickleTable) error { // nolint
	if len(data.Rows) > 1 {
		groupsToCreate := &messages.PickleStepArgument_PickleTable{
			Rows: make([]*messages.PickleStepArgument_PickleTable_PickleTableRow, 1, (len(data.Rows)-1)*2+1),
		}
		groupsToCreate.Rows[0] = &messages.PickleStepArgument_PickleTable_PickleTableRow{
			Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
				{Value: "id"}, {Value: "name"}, {Value: "description"}, {Value: "type"},
			},
		}
		head := data.Rows[0].Cells
		groupIDColumnNumber := -1
		loginColumnNumber := -1
		for number, cell := range head {
			if cell.Value == "group_id" {
				groupIDColumnNumber = number
				continue
			}
			if cell.Value == "login" {
				loginColumnNumber = number
				continue
			}
		}

		for i := 1; i < len(data.Rows); i++ {
			login := "null"
			if loginColumnNumber != -1 {
				login = data.Rows[i].Cells[loginColumnNumber].Value
			}

			if groupIDColumnNumber != -1 {
				groupsToCreate.Rows = append(groupsToCreate.Rows, &messages.PickleStepArgument_PickleTable_PickleTableRow{
					Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
						{Value: data.Rows[i].Cells[groupIDColumnNumber].Value}, {Value: login}, {Value: login}, {Value: "User"},
					},
				})
			}
		}

		if err := ctx.DBHasTable("groups", groupsToCreate); err != nil {
			return err
		}
	}

	return ctx.DBHasTable("users", data)
}

func (ctx *TestContext) DBGroupsAncestorsAreComputed() error { // nolint
	gormDB, err := database.Open(ctx.db())
	if err != nil {
		return err
	}

	err = database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
		return store.GroupGroups().After()
	})
	if err != nil {
		return err
	}

	ctx.dbTableData["groups_ancestors"] = &messages.PickleStepArgument_PickleTable{
		Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
			{Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
				{Value: "ancestor_group_id"}, {Value: "child_group_id"}, {Value: "expires_at"},
			}},
		},
	}

	var groupsAncestors []map[string]interface{}
	err = gormDB.Table("groups_ancestors").Select("ancestor_group_id, child_group_id, expires_at").
		Order("ancestor_group_id, child_group_id, expires_at").ScanIntoSliceOfMaps(&groupsAncestors).Error()

	if err != nil {
		return err
	}

	for _, row := range groupsAncestors {
		ctx.dbTableData["groups_ancestors"].Rows = append(ctx.dbTableData["groups_ancestors"].Rows,
			&messages.PickleStepArgument_PickleTable_PickleTableRow{
				Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
					{Value: row["ancestor_group_id"].(string)}, {Value: row["child_group_id"].(string)}, {Value: row["expires_at"].(string)},
				},
			})
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

func (ctx *TestContext) TableAtColumnValueShouldBeEmpty(tableName string, columnName, columnValues string) error { // nolint
	db := ctx.db()
	_, values, where := constructWhereForColumnValues(columnName, parseMultipleValuesString(columnValues), false)
	sqlRows, err := db.Query(fmt.Sprintf("SELECT 1 FROM %s %s LIMIT 1", tableName, where), values...) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = sqlRows.Close() }()
	if sqlRows.Next() {
		return fmt.Errorf("the table %q should be empty, but it is not", tableName)
	}

	return nil
}

func (ctx *TestContext) TableShouldBe(tableName string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	return ctx.tableAtColumnValueShouldBe(tableName, "", nil, false, data)
}

func (ctx *TestContext) TableShouldStayUnchanged(tableName string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &messages.PickleStepArgument_PickleTable{Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
			{Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{{Value: "1"}}}},
		}
	}
	return ctx.tableAtColumnValueShouldBe(tableName, "", nil, false, data)
}

func (ctx *TestContext) TableShouldStayUnchangedButTheRowWithColumnValue(tableName, columnName, columnValues string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &messages.PickleStepArgument_PickleTable{Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{}}
	}
	return ctx.tableAtColumnValueShouldBe(tableName, columnName, parseMultipleValuesString(columnValues), true, data)
}

func (ctx *TestContext) TableAtColumnValueShouldBe(tableName, columnName, columnValues string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	return ctx.tableAtColumnValueShouldBe(tableName, columnName, parseMultipleValuesString(columnValues), false, data)
}

func (ctx *TestContext) TableShouldNotContainColumnValue(tableName, columnName, columnValues string) error { // nolint
	return ctx.tableAtColumnValueShouldBe(tableName, columnName, parseMultipleValuesString(columnValues), false,
		&messages.PickleStepArgument_PickleTable{

			Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
				{Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{{Value: columnName}}}},
		})
}

func combinePickleTables(table1, table2 *messages.PickleStepArgument_PickleTable) *messages.PickleStepArgument_PickleTable {
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

	combinedTable := &messages.PickleStepArgument_PickleTable{}
	combinedTable.Rows = make([]*messages.PickleStepArgument_PickleTable_PickleTableRow, 0, len(table1.Rows)+len(table2.Rows)-1)

	header := &messages.PickleStepArgument_PickleTable_PickleTableRow{
		Cells: make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, 0, columnNumber),
	}
	for _, columnName := range combinedColumnNames {
		header.Cells = append(header.Cells, &messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{Value: columnName})
	}
	combinedTable.Rows = append(combinedTable.Rows, header)

	copyCellsIntoCombinedTable(table1, combinedColumnNames, table1FieldMap, combinedTable)
	copyCellsIntoCombinedTable(table2, combinedColumnNames, table2FieldMap, combinedTable)
	return combinedTable
}

func copyCellsIntoCombinedTable(sourceTable *messages.PickleStepArgument_PickleTable, combinedColumnNames []string,
	sourceTableFieldMap map[string]int, combinedTable *messages.PickleStepArgument_PickleTable) {
	for rowNum := 1; rowNum < len(sourceTable.Rows); rowNum++ {
		newRow := &messages.PickleStepArgument_PickleTable_PickleTableRow{
			Cells: make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, 0, len(combinedColumnNames)),
		}
		for _, column := range combinedColumnNames {
			var newCell *messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell
			if sourceColumnNumber, ok := sourceTableFieldMap[column]; ok {
				newCell = sourceTable.Rows[rowNum].Cells[sourceColumnNumber]
			}
			newRow.Cells = append(newRow.Cells, newCell)
		}
		combinedTable.Rows = append(combinedTable.Rows, newRow)
	}
}

func parseMultipleValuesString(valuesString string) []string {
	return strings.Split(valuesString, ",")
}

var columnNameRegexp = regexp.MustCompile(`^[a-zA-Z]\w*$`)

func (ctx *TestContext) tableAtColumnValueShouldBe(tableName, columnName string, columnValues []string, excludeValues bool, data *messages.PickleStepArgument_PickleTable) error { // nolint
	// For that, we build a SQL request with only the attributes we are interested about (those
	// for the test data table) and we convert them to string (in SQL) to compare to table value.
	// Expect 'null' string in the table to check for nullness

	db := ctx.db()

	var selects []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		dataTableColumnName := cell.Value
		if columnNameRegexp.MatchString(dataTableColumnName) {
			dataTableColumnName = database.QuoteName(dataTableColumnName)
		}
		selects = append(selects, dataTableColumnName)
	}

	columnValuesMap, values, where := constructWhereForColumnValues(columnName, columnValues, excludeValues)

	selectsJoined := strings.Join(selects, ", ")

	// exec sql
	query := fmt.Sprintf("SELECT %s FROM `%s` %s ORDER BY %s", selectsJoined, tableName, where, selectsJoined) // nolint: gosec
	sqlRows, err := db.Query(query, values...)
	if err != nil {
		return err
	}
	defer func() { _ = sqlRows.Close() }()
	dataCols := data.Rows[0].Cells
	columnIndex := -1
	for index, cell := range dataCols {
		if cell.Value == columnName {
			columnIndex = index
			break
		}
	}

	iDataRow := 1
	sqlCols, _ := sqlRows.Columns() // nolint: gosec
	for sqlRows.Next() {
		for excludeValues && iDataRow < len(data.Rows) && columnValuesMap[data.Rows[iDataRow].Cells[columnIndex].Value] {
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

	for excludeValues && iDataRow < len(data.Rows) && columnValuesMap[data.Rows[iDataRow].Cells[columnIndex].Value] {
		iDataRow++
	}

	// check that no row in the test data table has not been uncheck (if less rows in SQL result)
	if iDataRow < len(data.Rows) {
		return fmt.Errorf("there are less rows in the SQL results than expected")
	}
	return nil
}

func constructWhereForColumnValues(columnName string, columnValues []string, excludeValues bool) (
	columnValuesMap map[string]bool, values []interface{}, where string) {
	columnValuesMap = make(map[string]bool, len(columnValues))
	for _, value := range columnValues {
		columnValuesMap[value] = true
	}
	values = make([]interface{}, 0, len(columnValues))
	for value := range columnValuesMap {
		values = append(values, value)
	}
	// define 'where' condition if needed
	where = ""
	if len(columnValues) > 0 {
		questionMarks := "?" + strings.Repeat(", ?", len(columnValues)-1)
		if excludeValues {
			where = fmt.Sprintf(" WHERE %s NOT IN (%s) ", columnName, questionMarks) // #nosec
		} else {
			where = fmt.Sprintf(" WHERE %s IN (%s) ", columnName, questionMarks) // #nosec
		}
	}
	return columnValuesMap, values, where
}

func (ctx *TestContext) DbTimeNow(timeStrRaw string) error { // nolint
	MockDBTime(timeStrRaw)
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
