package testhelpers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/godog/gherkin"
	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/jinzhu/gorm"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
	log "github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type dbquery struct {
	sql    string
	values []interface{}
}

type addedDBIndex struct {
	Table string
	Index string
}

// TestContext implements context for tests
type TestContext struct {
	// nolint
	application      *app.Application // do NOT call it directly, use `app()`
	userID           int64            // userID that will be used for the next requests
	featureQueries   []dbquery
	lastResponse     *http.Response
	lastResponseBody string
	inScenario       bool
	dbTableData      map[string]*gherkin.DataTable
	addedDBIndices   []*addedDBIndex
}

var db *sql.DB

func (ctx *TestContext) SetupTestContext(data interface{}) { // nolint
	scenario := data.(*gherkin.Scenario)
	log.WithField("type", "test").Infof("Starting test scenario: %s", scenario.Name)
	ctx.setupApp()
	ctx.userID = 999 // the default for the moment
	ctx.lastResponse = nil
	ctx.lastResponseBody = ""
	ctx.inScenario = true
	ctx.dbTableData = make(map[string]*gherkin.DataTable)

	// reset the seed to get predictable results on PRNG for tests
	rand.Seed(1)

	// fix the current time
	currentTime := time.Now()
	monkey.Patch(time.Now, func() time.Time { return currentTime })

	err := ctx.initDB()
	if err != nil {
		fmt.Println("Unable to empty db")
		panic(err)
	}
}

func (ctx *TestContext) setupApp() {
	var err error
	ctx.tearDownApp()
	ctx.application, err = app.New()
	if err != nil {
		fmt.Println("Unable to load app")
		panic(err)
	}
}

func (ctx *TestContext) tearDownApp() {
	if ctx.application != nil {
		_ = ctx.application.Database.Close()
	}
	ctx.application = nil
}

func (ctx *TestContext) ScenarioTeardown(interface{}, error) { // nolint
	monkey.UnpatchAll()

	db, err := gorm.Open("mysql", ctx.db())
	if err != nil {
		panic(err)
	}

	for _, indexDefinition := range ctx.addedDBIndices {
		if oneErr := db.Table(indexDefinition.Table).RemoveIndex(indexDefinition.Index).Error; oneErr != nil {
			_ = db.AddError(oneErr) // nolint: gosec
		}
	}
	if db.Error != nil {
		panic(db.Error)
	}

	ctx.tearDownApp()
}

func testRequest(ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string, error) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		return nil, "", err
	}

	// set a dummy auth cookie
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: "dummy"})

	// execute the query
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	defer func() { /* #nosec */ _ = resp.Body.Close() }()

	return resp, string(respBody), nil
}

func (ctx *TestContext) setupAuthProxyServer() *httptest.Server {
	// set the auth proxy server up
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		dataJSON := fmt.Sprintf(`{"userID": %d, "error":""}`, ctx.userID)
		_, _ = w.Write([]byte(dataJSON)) // nolint
	}))

	// put the backend URL into the config
	backendURL, _ := url.Parse(backend.URL) // nolint
	ctx.application.Config.Auth.ProxyURL = backendURL.String()

	return backend
}

func (ctx *TestContext) db() *sql.DB {
	if db == nil {
		conf := ctx.application.Config
		var err error
		db, err = sql.Open("mysql", conf.Database.Connection.FormatDSN())
		if err != nil {
			fmt.Println("Unable to connect to the database: ", err)
			os.Exit(1)
		}
	}
	return db
}

// nolint: gosec
func (ctx *TestContext) emptyDB() error {

	db := ctx.db()

	dbName := ctx.application.Config.Database.Connection.DBName
	rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                         FROM   information_schema.tables
                         WHERE  table_type   = 'BASE TABLE'
                           AND  table_schema = '` + dbName + `'
                           AND  table_name  != 'gorp_migrations'`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tableName string
		if scanErr := rows.Scan(&tableName); scanErr != nil {
			return scanErr
		}
		// DELETE is MUCH faster than TRUNCATE on empty tables
		_, err := db.Exec("DELETE FROM " + tableName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *TestContext) initDB() error {
	err := ctx.emptyDB()
	if err != nil {
		return err
	}
	db := ctx.db()

	for _, query := range ctx.featureQueries {
		_, err := db.Exec(query.sql, query.values)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *TestContext) iSendrequestGeneric(method, path, reqBody string) error {
	// app server
	testServer := httptest.NewServer(ctx.application.HTTPHandler)
	defer testServer.Close()

	// auth proxy server
	authProxyServer := ctx.setupAuthProxyServer()
	defer authProxyServer.Close()

	// do request
	response, body, err := testRequest(testServer, method, path, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	ctx.lastResponse = response
	ctx.lastResponseBody = body

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

var prepareValRegexp = regexp.MustCompile(`^\s*([\w]+)\s*\(\s*(.*)\)\s*$`)

func prepareVal(input string) string {
	if match := prepareValRegexp.FindStringSubmatch(input); len(match) == 3 && match[1] == "relativeTime" {
		duration, err := time.ParseDuration(match[2])
		if err != nil {
			panic(err)
		}
		return time.Now().UTC().Add(duration).Format(time.RFC3339)
	}
	return input
}

/** Steps **/

func (ctx *TestContext) DBHasTable(tableName string, data *gherkin.DataTable) error { // nolint
	db := ctx.db()

	head := data.Rows[0].Cells
	fields := make([]string, 0, len(head))
	marks := make([]string, 0, len(head))
	for _, cell := range head {
		fields = append(fields, cell.Value)
		marks = append(marks, "?")
	}
	query := "INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES(" + strings.Join(marks, ", ") + ")" // nolint: gosec
	for i := 1; i < len(data.Rows); i++ {
		var vals []interface{}
		for _, cell := range data.Rows[i].Cells {
			cell.Value = prepareVal(cell.Value)
			vals = append(vals, dbDataTableValue(cell.Value))
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

func (ctx *TestContext) RunFallbackServer() error { // nolint
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Got-Query", r.URL.Path)
	}))
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		return err
	}

	_ = os.Setenv("ALGOREA_REVERSEPROXY.SERVER", backendURL.String()) // nolint
	ctx.setupApp()
	return nil
}

func (ctx *TestContext) IAmUserWithID(id int64) error { // nolint
	ctx.userID = id
	return nil
}

func (ctx *TestContext) TimeNow(timeStr string) error { // nolint
	testTime, err := time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		monkey.Patch(time.Now, func() time.Time { return testTime })
	}
	return err
}

func (ctx *TestContext) ISendrequestToWithBody(method string, path string, body *gherkin.DocString) error { // nolint
	return ctx.iSendrequestGeneric(method, path, body.Content)
}

func (ctx *TestContext) ISendrequestTo(method string, path string) error { // nolint
	return ctx.iSendrequestGeneric(method, path, "")
}

func (ctx *TestContext) ItShouldBeAJSONArrayWithEntries(count int) error { // nolint
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s\nData:%v", err, ctx.lastResponseBody)
	}

	if count != len(objmap) {
		return fmt.Errorf("the result does not have the expected length. Expected: %d, received: %d", count, len(objmap))
	}

	return nil
}

func (ctx *TestContext) TheResponseCodeShouldBe(code int) error { // nolint
	if code != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected http response code: %d, actual is: %d. \n Data: %s", code, ctx.lastResponse.StatusCode, ctx.lastResponseBody)
	}
	return nil
}

var jsonPrepareRegexp = regexp.MustCompile(`{\s*(\w+)\[(\d+)]\[(\w+)]}`)

func (ctx *TestContext) TheResponseBodyShouldBeJSON(body *gherkin.DocString) (err error) { // nolint
	var expected, actual []byte
	var exp, act interface{}

	// verify the content type
	if err = ValidateJSONContentType(ctx.lastResponse); err != nil {
		return
	}

	expectedBody := body.Content
	for match := jsonPrepareRegexp.FindStringSubmatch(expectedBody); match != nil; match = jsonPrepareRegexp.FindStringSubmatch(expectedBody) {
		gherkinTable := ctx.dbTableData[match[1]]
		neededColumnNumber := -1
		for columnNumber, cell := range gherkinTable.Rows[0].Cells {
			if cell.Value == match[3] {
				neededColumnNumber = columnNumber
				break
			}
		}
		if neededColumnNumber == -1 {
			panic(fmt.Errorf("cannot find column %q in table %q", match[3], match[1]))
		}
		rowNumber, conversionErr := strconv.Atoi(match[2])
		if conversionErr != nil {
			panic(conversionErr)
		}
		expectedBody = strings.Replace(expectedBody, match[0], gherkinTable.Rows[rowNumber].Cells[neededColumnNumber].Value, -1)
	}

	// re-encode expected response
	if err = json.Unmarshal([]byte(expectedBody), &exp); err != nil {
		return
	}
	if expected, err = json.MarshalIndent(exp, "", "\t"); err != nil {
		return
	}

	// re-encode actual response too
	if err = json.Unmarshal([]byte(ctx.lastResponseBody), &act); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s -- Data: %v", err, ctx.lastResponseBody)
	}
	if actual, err = json.MarshalIndent(act, "", "\t"); err != nil {
		return
	}

	sExpected := string(expected)
	sActual := string(actual)

	if sExpected != sActual {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{ // nolint: gosec
			A:        difflib.SplitLines(sExpected),
			B:        difflib.SplitLines(sActual),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})

		return fmt.Errorf(
			"expected JSON does not match actual.\n     Diff:\n%s",
			diff,
		)
	}
	return err
}

func (ctx *TestContext) TheResponseHeaderShouldBe(headerName string, headerValue string) (err error) { // nolint
	if ctx.lastResponse.Header.Get(headerName) != headerValue {
		return fmt.Errorf("headers %s different from expected. Expected: %s, got: %s",
			headerName, headerValue, ctx.lastResponse.Header.Get(headerName))
	}
	return nil
}

func (ctx *TestContext) TheResponseErrorMessageShouldContain(s string) (err error) { // nolint

	errorResp := service.ErrorResponse{}
	// decode response
	if err = json.Unmarshal([]byte(ctx.lastResponseBody), &errorResp); err != nil {
		return fmt.Errorf("unable to decode the response as JSON: %s -- Data: %v", err, ctx.lastResponseBody)
	}
	if !strings.Contains(errorResp.ErrorText, s) {
		return fmt.Errorf("cannot find expected `%s` in error text: `%s`", s, errorResp.ErrorText)
	}

	return nil
}

func (ctx *TestContext) TheResponseShouldBe(kind string) error { // nolint
	var expectedCode int
	switch kind {
	case "updated":
		expectedCode = 200
	case "created":
		expectedCode = 201
	default:
		return fmt.Errorf("unknown response kind: %q", kind)
	}
	if err := ctx.TheResponseCodeShouldBe(expectedCode); err != nil {
		return err
	}
	if err := ctx.TheResponseBodyShouldBeJSON(&gherkin.DocString{
		Content: `
		{
			"message": "` + kind + `",
			"success": true
		}`}); err != nil {
		return err
	}
	return nil
}

func (ctx *TestContext) TableShouldBe(tableName string, data *gherkin.DataTable) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, nil, false, data)
}

func (ctx *TestContext) TableShouldStayUnchanged(tableName string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &gherkin.DataTable{Rows: []*gherkin.TableRow{}}
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

func (ctx *TestContext) TableAtIDShouldBe(tableName string, ids string, data *gherkin.DataTable) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, parseMultipleIDString(ids), false, data)
}

func (ctx *TestContext) TableShouldNotContainID(tableName string, ids string) error { // nolint
	return ctx.tableAtIDShouldBe(tableName, parseMultipleIDString(ids), false,
		&gherkin.DataTable{Rows: []*gherkin.TableRow{{Cells: []*gherkin.TableCell{{Value: "ID"}}}}})
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
			dataValue := prepareVal(dataCell.Value)
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

func (ctx *TestContext) TheGeneratedGroupPasswordIs(generatedPassword string) error { // nolint
	monkey.Patch(groups.GenerateGroupPassword, func() (string, error) { return generatedPassword, nil })
	return nil
}

var multipleStringsRegexp = regexp.MustCompile(`^((?:\s*,\s*)?"([^"]*)")`)

func (ctx *TestContext) TheGeneratedGroupPasswordsAre(generatedPasswords string) error { // nolint
	currentIndex := 0
	monkey.Patch(groups.GenerateGroupPassword, func() (string, error) {
		currentIndex++
		password := multipleStringsRegexp.FindStringSubmatch(generatedPasswords)
		if password == nil {
			return "", errors.New("not enough generated passwords")
		}
		generatedPasswords = generatedPasswords[len(password[1]):]
		return password[2], nil
	})
	return nil
}
