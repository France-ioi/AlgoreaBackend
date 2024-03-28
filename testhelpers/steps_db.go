//go:build !prod

package testhelpers

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/messages-go/v10"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

type rowTransformation int

const (
	unchanged rowTransformation = iota + 1
	changed
	deleted
)

const (
	UserGroupID = "group_id"
	UserLogin   = "login"
)

func (ctx *TestContext) DBHasTable(table string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	db = ctx.db()

	if len(data.Rows) > 1 {
		referenceColumnIndex := -1
		head := data.Rows[0].Cells
		fields := make([]string, 0, len(head))
		marks := make([]string, 0, len(head))

		for i, cell := range head {
			if cell.Value == "@reference" {
				referenceColumnIndex = i
			}

			fields = append(fields, database.QuoteName(cell.Value))
			marks = append(marks, "?")
		}

		marksString := "(" + strings.Join(marks, ", ") + ")"
		finalMarksString := marksString
		if len(data.Rows) > 2 {
			finalMarksString = strings.Repeat(marksString+", ", len(data.Rows)-2) + finalMarksString
		}
		query := "INSERT INTO " + database.QuoteName(table) + // nolint: gosec
			" (" + strings.Join(fields, ", ") + ") VALUES " + finalMarksString
		vals := make([]interface{}, 0, (len(data.Rows)-1)*len(head))
		for i := 1; i < len(data.Rows); i++ {
			for j, cell := range data.Rows[i].Cells {
				if j != referenceColumnIndex {
					var err error
					if cell.Value, err = ctx.preprocessString(cell.Value); err != nil {
						return err
					}
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

	if ctx.dbTableData[table] == nil {
		ctx.dbTableData[table] = data
	} else if len(data.Rows) > 1 {
		ctx.dbTableData[table] = combinePickleTables(ctx.dbTableData[table], data)
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
			if cell.Value == UserGroupID {
				groupIDColumnNumber = number
				continue
			}
			if cell.Value == UserLogin {
				loginColumnNumber = number
				continue
			}
		}

		for i := 1; i < len(data.Rows); i++ {
			login := tableValueNull
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

// saveTableFromDatabase saves the content of a table in database for some columns, in order to later check if the table
// had changed after some manipulations.
func (ctx *TestContext) saveTableFromDatabase(gormDB *database.DB, table string, columns []string) error {
	headerCells := make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, len(columns))
	for i, column := range columns {
		headerCells[i] = &messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
			Value: column,
		}
	}

	ctx.dbTableData[table] = &messages.PickleStepArgument_PickleTable{
		Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
			{Cells: headerCells},
		},
	}

	var rows []map[string]interface{}
	err := gormDB.Table(table).Select(strings.Join(columns, ",")).
		Order(strings.Join(columns, ",")).ScanIntoSliceOfMaps(&rows).Error()
	if err != nil {
		return err
	}

	for _, row := range rows {
		rowCells := make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, len(columns))
		for j, column := range columns {
			rowCells[j] = &messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{
				Value: row[column].(string),
			}
		}

		ctx.dbTableData[table].Rows = append(ctx.dbTableData[table].Rows, &messages.PickleStepArgument_PickleTable_PickleTableRow{
			Cells: rowCells,
		})
	}

	return nil
}

// DBGroupsAncestorsAreComputed computes the groups_ancestors table.
func (ctx *TestContext) DBGroupsAncestorsAreComputed() error {
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

	err = ctx.saveTableFromDatabase(gormDB, "groups_ancestors", []string{
		"ancestor_group_id",
		"child_group_id",
		"expires_at",
	})
	if err != nil {
		return err
	}

	return nil
}

// DBItemsAncestorsAndPermissionsAreComputed computes the items_ancestors and permissions_generated tables.
func (ctx *TestContext) DBItemsAncestorsAndPermissionsAreComputed() error {
	gormDB, err := database.Open(ctx.db())
	if err != nil {
		return err
	}

	err = database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
		// We can consider keeping foreign_key_checks,
		// but it'll break all tests that didn't define items while having permissions.
		store.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer store.Exec("SET FOREIGN_KEY_CHECKS=1")

		return store.ItemItems().After()
	})
	if err != nil {
		return err
	}

	err = ctx.saveTableFromDatabase(gormDB, "items_ancestors", []string{
		"ancestor_item_id",
		"child_item_id",
	})
	if err != nil {
		return err
	}

	err = ctx.saveTableFromDatabase(gormDB, "permissions_generated", []string{
		"group_id",
		"item_id",
		"can_view_generated",
		"can_grant_view_generated",
		"can_watch_generated",
		"can_edit_generated",
		"is_owner_generated",
		"can_view_generated_value",
		"can_grant_view_generated_value",
		"can_watch_generated_value",
		"can_edit_generated_value",
	})
	if err != nil {
		return err
	}

	return nil
}

func (ctx *TestContext) TableShouldBeEmpty(table string) error { //nolint
	db = ctx.db()
	sqlRows, err := db.Query(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table)) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlRows.Close()

		if sqlRows.Err() != nil {
			panic(sqlRows.Err())
		}
	}()
	if sqlRows.Next() {
		return fmt.Errorf("the table %q should be empty, but it is not", table)
	}

	return nil
}

func (ctx *TestContext) TableAtColumnValueShouldBeEmpty(table string, column, valuesStr string) error { //nolint
	values := parseMultipleValuesString(valuesStr)

	db = ctx.db()
	where, parameters := constructWhereForColumnValues([]string{column}, values, true)
	sqlRows, err := db.Query(fmt.Sprintf("SELECT 1 FROM %s %s LIMIT 1", table, where), parameters...) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlRows.Close()

		if sqlRows.Err() != nil {
			panic(sqlRows.Err())
		}
	}()
	if sqlRows.Next() {
		return fmt.Errorf("the table %q should be empty, but it is not", table)
	}

	return nil
}

func (ctx *TestContext) TableShouldBe(table string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	return ctx.tableAtColumnValueShouldBe(table, []string{""}, nil, unchanged, data)
}

func (ctx *TestContext) TableShouldStayUnchanged(table string) error { //nolint
	data := ctx.dbTableData[table]
	if data == nil {
		data = &messages.PickleStepArgument_PickleTable{
			Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
				{Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{{Value: "1"}}},
			},
		}
	}
	return ctx.tableAtColumnValueShouldBe(table, []string{""}, nil, unchanged, data)
}

func (ctx *TestContext) TableShouldStayUnchangedButTheRowWithColumnValue(table, column, values string) error { //nolint
	data := ctx.dbTableData[table]
	if data == nil {
		data = &messages.PickleStepArgument_PickleTable{Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{}}
	}
	return ctx.tableAtColumnValueShouldBe(table, []string{column}, parseMultipleValuesString(values), changed, data)
}

// TableShouldStayUnchangedButTheRowsWithColumnValueShouldBeDeleted checks for row deletion.
func (ctx *TestContext) TableShouldStayUnchangedButTheRowsWithColumnValueShouldBeDeleted(table, columns, values string) error {
	data := ctx.dbTableData[table]
	if data == nil {
		data = &messages.PickleStepArgument_PickleTable{Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{}}
	}

	return ctx.tableAtColumnValueShouldBe(table, parseMultipleValuesString(columns), parseMultipleValuesString(values), deleted, data)
}

func (ctx *TestContext) TableAtColumnValueShouldBe(table, column, values string, data *messages.PickleStepArgument_PickleTable) error { // nolint
	return ctx.tableAtColumnValueShouldBe(
		table,
		[]string{column},
		parseMultipleValuesString(ctx.replaceReferencesByIDs(values)),
		unchanged,
		data,
	)
}

func (ctx *TestContext) TableShouldNotContainColumnValue(table, column, values string) error { //nolint
	return ctx.tableAtColumnValueShouldBe(table, []string{column}, parseMultipleValuesString(ctx.replaceReferencesByIDs(values)), unchanged,
		&messages.PickleStepArgument_PickleTable{
			Rows: []*messages.PickleStepArgument_PickleTable_PickleTableRow{
				{Cells: []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{{Value: column}}},
			},
		})
}

func combinePickleTables(table1, table2 *messages.PickleStepArgument_PickleTable) *messages.PickleStepArgument_PickleTable {
	table1FieldMap := map[string]int{}
	combinedFieldMap := map[string]bool{}
	columnNumber := len(table1.Rows[0].Cells)
	combinedcolumns := make([]string, 0, columnNumber+len(table2.Rows[0].Cells))
	for index, cell := range table1.Rows[0].Cells {
		table1FieldMap[cell.Value] = index
		combinedFieldMap[cell.Value] = true
		combinedcolumns = append(combinedcolumns, cell.Value)
	}
	table2FieldMap := map[string]int{}
	for index, cell := range table2.Rows[0].Cells {
		table2FieldMap[cell.Value] = index
		// only add a column if it hasn't been met in table1
		if !combinedFieldMap[cell.Value] {
			combinedFieldMap[cell.Value] = true
			columnNumber++
			combinedcolumns = append(combinedcolumns, cell.Value)
		}
	}

	combinedTable := &messages.PickleStepArgument_PickleTable{}
	combinedTable.Rows = make([]*messages.PickleStepArgument_PickleTable_PickleTableRow, 0, len(table1.Rows)+len(table2.Rows)-1)

	header := &messages.PickleStepArgument_PickleTable_PickleTableRow{
		Cells: make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, 0, columnNumber),
	}
	for _, column := range combinedcolumns {
		header.Cells = append(header.Cells, &messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell{Value: column})
	}
	combinedTable.Rows = append(combinedTable.Rows, header)

	copyCellsIntoCombinedTable(table1, combinedcolumns, table1FieldMap, combinedTable)
	copyCellsIntoCombinedTable(table2, combinedcolumns, table2FieldMap, combinedTable)
	return combinedTable
}

func copyCellsIntoCombinedTable(sourceTable *messages.PickleStepArgument_PickleTable, combinedcolumns []string,
	sourceTableFieldMap map[string]int, combinedTable *messages.PickleStepArgument_PickleTable,
) {
	for rowNum := 1; rowNum < len(sourceTable.Rows); rowNum++ {
		newRow := &messages.PickleStepArgument_PickleTable_PickleTableRow{
			Cells: make([]*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell, 0, len(combinedcolumns)),
		}
		for _, column := range combinedcolumns {
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

var columnRegexp = regexp.MustCompile(`^[a-zA-Z]\w*$`)

func (ctx *TestContext) tableAtColumnValueShouldBe(table string, columns, values []string,
	rowTransformation rowTransformation, data *messages.PickleStepArgument_PickleTable,
) error { // nolint
	// For that, we build a SQL request with only the attributes we are interested about (those
	// for the test data table) and we convert them to string (in SQL) to compare to table value.
	// Expect 'null' string in the table to check for nullness
	// Expect 'null' string in the table to check for nullness

	if rowTransformation == deleted {
		nbRows, err := ctx.getNbRowsMatching(table, columns, values)
		if err != nil {
			return err
		}
		if nbRows > 0 {
			return fmt.Errorf("found %d rows that should have been deleted", nbRows)
		}
	}

	tableColumns := getColumnNamesFromData(data)

	// request for "unchanged": WHERE IN...
	// request for "changed" or "deleted": WHERE NOT IN...
	whereIn := rowTransformation == unchanged
	sqlRows, closer, err := ctx.getSQLRowsMatching(table, tableColumns, columns, values, whereIn)
	if err != nil {
		return err
	}
	defer closer()

	err = ctx.dataTableMatchesSQLRows(data, sqlRows, rowTransformation, tableColumns, columns, values)
	if err != nil {
		return err
	}

	return nil
}

// dataTableMatchesSQLRows checks whether the provided data table matches the database rows result.
func (ctx *TestContext) dataTableMatchesSQLRows(data *messages.PickleStepArgument_PickleTable, sqlRows *sql.Rows,
	rowTransformation rowTransformation, tableColumns, columns, values []string,
) error {
	iDataRow := 1
	columnIndexes := getColumnIndexes(data, columns)
	for sqlRows.Next() {
		for shouldSkipRow(data, iDataRow, columnIndexes, values, rowTransformation) {
			iDataRow++
		}
		if iDataRow >= len(data.Rows) {
			return fmt.Errorf("there are more rows in the SQL results than expected. expected: %d", len(data.Rows)-1)
		}

		// We need pointers to differentiate null values
		sqlRowValues, err := getStringPtrFromSQLRow(sqlRows, len(tableColumns))
		if err != nil {
			return err
		}

		err = ctx.dataRowMatchesSQLRow(data.Rows[iDataRow], sqlRowValues, tableColumns, iDataRow)
		if err != nil {
			return err
		}

		iDataRow++
	}
	for shouldSkipRow(data, iDataRow, columnIndexes, values, rowTransformation) {
		iDataRow++
	}

	// check that there are no rows in the test data table left for checking (this means there are fewer rows in the SQL result)
	if iDataRow < len(data.Rows) {
		return fmt.Errorf("there are fewer rows in the SQL result than expected")
	}

	return nil
}

// dataRowMatchesSQLRow checks that a data row matches a row from database.
func (ctx *TestContext) dataRowMatchesSQLRow(dataRow *messages.PickleStepArgument_PickleTable_PickleTableRow,
	values []*string, tableColumns []string, rowIndex int,
) error {
	// checking that all columns of the test data table match the SQL row
	for colIndex, dataCell := range dataRow.Cells {
		if dataCell == nil {
			continue
		}

		dataValue, err := ctx.preprocessString(dataCell.Value)
		if err != nil {
			return err
		}

		value := values[colIndex]
		if value == nil {
			value = pTableValueNull
		}

		if (dataValue == tableValueTrue && *value == "1") || (dataValue == tableValueFalse && *value == "0") {
			continue
		}

		if dataValue != *value {
			return fmt.Errorf("not matching expected value at row %d, col %s, expected '%s', got: '%v'",
				rowIndex, tableColumns[colIndex], dataValue, *value)
		}
	}

	return nil
}

// getColumnNamesFromData gets the column names from the data table.
func getColumnNamesFromData(data *messages.PickleStepArgument_PickleTable) (columns []string) {
	// the first row contains the column names
	headerColumns := data.Rows[0].Cells
	for _, cell := range headerColumns {
		dataTablecolumn := cell.Value
		if columnRegexp.MatchString(dataTablecolumn) {
			dataTablecolumn = database.QuoteName(dataTablecolumn)
		}

		columns = append(columns, dataTablecolumn)
	}

	return columns
}

// getStringPtrFromSQLRow gets a slice of string pointers from a SQL row.
func getStringPtrFromSQLRow(sqlRows *sql.Rows, length int) ([]*string, error) {
	// Create a slice of values and a second slice with pointers to each item.
	rowValues := make([]*string, length)
	rowValPtr := make([]interface{}, length)
	for i := range rowValues {
		rowValPtr[i] = &rowValues[i]
	}

	// Scan the result into the column pointers...
	err := sqlRows.Scan(rowValPtr...)

	return rowValues, err
}

// getColumnIndexes gets the indices of the columns referenced by columns.
func getColumnIndexes(data *messages.PickleStepArgument_PickleTable, columns []string) []int {
	// the first row contains the column names
	headerColumns := data.Rows[0].Cells

	columnIndexes := make([]int, len(columns))
	for i := range columnIndexes {
		columnIndexes[i] = -1
	}
	for headerColumnIndex, headerColumn := range headerColumns {
		for columnIndex, column := range columns {
			if headerColumn.Value == column {
				columnIndexes[columnIndex] = headerColumnIndex
				break
			}
		}
	}

	return columnIndexes
}

// getSQLRowsMatching returns the rows that matches (if whereIn) or not (if !whereIn) one of filterColumns at any filterValues.
func (ctx *TestContext) getSQLRowsMatching(table string, columns, filterColumns, filterValues []string, whereIn bool) (
	*sql.Rows, func(), error,
) {
	db = ctx.db()

	selectsJoined := strings.Join(columns, ", ")

	where, parameters := constructWhereForColumnValues(filterColumns, filterValues, whereIn)

	// exec sql
	query := fmt.Sprintf("SELECT %s FROM `%s` %s ORDER BY %s", selectsJoined, table, where, selectsJoined) //nolint: gosec
	sqlRows, err := db.Query(query, parameters...)

	closer := func() { _ = sqlRows.Close() }
	return sqlRows, closer, err
}

// getNbRowsMatching returns how many rows matches one of values at any column.
func (ctx *TestContext) getNbRowsMatching(table string, columns, values []string) (int, error) {
	db = ctx.db()

	// check that the rows are not present anymore
	where, parameters := constructWhereForColumnValues(columns, values, true)

	// exec sql
	var nbRows int
	selectValuesInQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s` %s", table, where) //nolint: gosec
	err := db.QueryRow(selectValuesInQuery, parameters...).Scan(&nbRows)

	return nbRows, err
}

func shouldSkipRow(data *messages.PickleStepArgument_PickleTable, rowIndex int, columnIndexes []int,
	values []string, rowTransformation rowTransformation,
) bool {
	return rowTransformation != unchanged &&
		rowIndex < len(data.Rows) &&
		rowMatchesColumnValues(data.Rows[rowIndex], columnIndexes, values)
}

// rowMatchesColumnValues checks whether a column matches some values at some rows
// we do an OR operation, thus returning if any column is match one of the values.
func rowMatchesColumnValues(row *messages.PickleStepArgument_PickleTable_PickleTableRow, columnIndexes []int, values []string) bool {
	// Both loops should contain 1 or 2 elements only
	for _, columnIndex := range columnIndexes {
		for _, value := range values {
			if row.Cells[columnIndex].Value == value {
				return true
			}
		}
	}

	return false
}

// constructWhereForColumnValues construct the WHERE part of a query matching column with values
// note: the same values are checked for every column
func constructWhereForColumnValues(columns, values []string, whereIn bool) (
	where string, parameters []interface{},
) {
	if len(values) > 0 {
		questionMarks := "?" + strings.Repeat(", ?", len(values)-1)

		isFirstCondition := true
		for _, column := range columns {
			if isFirstCondition {
				where += " WHERE "
			} else {
				where += " OR "
			}
			isFirstCondition = false

			if whereIn {
				where += fmt.Sprintf(" %s IN (%s) ", column, questionMarks) // #nosec
			} else {
				where += fmt.Sprintf(" %s NOT IN (%s) ", column, questionMarks) // #nosec
			}

			for _, value := range values {
				parameters = append(parameters, value)
			}
		}
	}

	return where, parameters
}

func (ctx *TestContext) DbTimeNow(timeStrRaw string) error { //nolint
	MockDBTime(timeStrRaw)
	return nil
}

const (
	tableValueFalse = "false"
	tableValueTrue  = "true"
	tableValueNull  = "null"
)

var (
	tableValueNullVar = tableValueNull
	pTableValueNull   = &tableValueNullVar
)

// dbDataTableValue converts a string value that we can find the db seeding table to a valid type for the db
// e.g., the string "null" means the SQL `NULL`.
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
