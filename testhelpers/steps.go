package testhelpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/godog/gherkin"
	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app"
	log "github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type dbquery struct {
	sql    string
	values []interface{}
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
}

const (
	noID int64 = math.MinInt64
)

func (ctx *TestContext) SetupTestContext(data interface{}) { // nolint
	scenario := data.(*gherkin.Scenario)
	log.WithField("type", "test").Infof("Starting test scenario: %s", scenario.Name)
	ctx.application = nil
	ctx.userID = 999 // the default for the moment
	ctx.lastResponse = nil
	ctx.lastResponseBody = ""
	ctx.inScenario = true
	ctx.dbTableData = make(map[string]*gherkin.DataTable)
}

func (ctx *TestContext) ScenarioTeardown(interface{}, error) { // nolint
}

func (ctx *TestContext) app() *app.Application {

	if ctx.application == nil {
		var err error
		ctx.application, err = app.New()
		if err != nil {
			fmt.Println("Unable to load app")
			panic(err)
		}
		// reset the seed to get predictable results on PRNG for tests
		rand.Seed(1)

		err = ctx.initDB()
		if err != nil {
			fmt.Println("Unable to empty db")
			panic(err)
		}
	}
	return ctx.application
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
	ctx.app().Config.Auth.ProxyURL = backendURL.String()

	return backend
}

func (ctx *TestContext) db() *sql.DB {
	conf := ctx.app().Config
	conn, err := sql.Open("mysql", conf.Database.Connection.FormatDSN())
	if err != nil {
		fmt.Println("Unable to connect to the database: ", err)
		os.Exit(1)
	}
	return conn
}

// nolint: gosec
func (ctx *TestContext) emptyDB() error {

	db := ctx.db()
	defer func() { _ = db.Close() }()

	dbName := ctx.app().Config.Database.Connection.DBName
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
		if err = rows.Scan(&tableName); err != nil {
			return err
		}
		_, err = db.Exec("TRUNCATE TABLE " + tableName)
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
	defer func() { /* #nosec */ _ = db.Close() }()

	for _, query := range ctx.featureQueries {
		_, err := db.Exec(query.sql, query.values)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *TestContext) iSendrequestGeneric(method string, path string, reqBody string) error {
	// app server
	testServer := httptest.NewServer(ctx.app().HTTPHandler)
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

// dbDataTableValue converts a string value that we can find the db seeding table to a valid type for the db
// e.g., the string "null" means the SQL `NULL`
func dbDataTableValue(input string) interface{} {
	switch input {
	case "false":
		return false
	case "true":
		return true
	case "null":
		return nil
	default:
		return input
	}
}

/** Steps **/

func (ctx *TestContext) DBHasTable(tableName string, data *gherkin.DataTable) error { // nolint

	db := ctx.db()
	defer func() { /* #nosec */ _ = db.Close() }()

	var fields []string
	var marks []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		fields = append(fields, cell.Value)
		marks = append(marks, "?")
	}
	query := "INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES(" + strings.Join(marks, ", ") + ")" // nolint: gosec
	for i := 1; i < len(data.Rows); i++ {
		var vals []interface{}
		for _, cell := range data.Rows[i].Cells {
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

	if len(data.Rows) > 1 {
		if ctx.dbTableData[tableName] != nil {
			ctx.dbTableData[tableName] = combineGherkinTables(ctx.dbTableData[tableName], data)
		} else {
			ctx.dbTableData[tableName] = data
		}
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
	viper.Set("ReverseProxy.Server", backendURL.String())
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

func (ctx *TestContext) TheResponseBodyShouldBeJSON(body *gherkin.DocString) (err error) { // nolint
	var expected, actual []byte
	var exp, act interface{}

	// re-encode expected response
	if err = json.Unmarshal([]byte(body.Content), &exp); err != nil {
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
			A:        difflib.SplitLines(string(expected)),
			B:        difflib.SplitLines(string(actual)),
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
	return
}

func (ctx *TestContext) TheResponseHeaderShouldBe(headerName string, headerValue string) (err error) { // nolint
	if ctx.lastResponse.Header.Get(headerName) != headerValue {
		return fmt.Errorf("headers %s different from expected. Expected: %s, got: %s", headerName, headerValue, ctx.lastResponse.Header.Get(headerName))
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
	return ctx.TableAtIDShouldBe(tableName, noID, data)
}

func (ctx *TestContext) TableShouldStayUnchanged(tableName string) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &gherkin.DataTable{Rows: []*gherkin.TableRow{}}
	}
	return ctx.TableAtIDShouldBe(tableName, noID, data)
}

func (ctx *TestContext) TableShouldStayUnchangedButTheRowWithID(tableName string, id int64) error { // nolint
	data := ctx.dbTableData[tableName]
	if data == nil {
		data = &gherkin.DataTable{Rows: []*gherkin.TableRow{}}
	}
	idColumnIndex := -1
	for index, cell := range data.Rows[0].Cells {
		if cell.Value == "ID" {
			idColumnIndex = index
			break
		}
	}

	idStringValue := strconv.FormatInt(id, 10)
	newData := &gherkin.DataTable{Rows: make([]*gherkin.TableRow, 0, len(data.Rows))}
	for index, row := range data.Rows {
		if index == 0 || idColumnIndex < 0 ||
			row.Cells[idColumnIndex] == nil || row.Cells[idColumnIndex].Value != idStringValue {
			newData.Rows = append(newData.Rows, row)
		}
	}
	return ctx.TableAtIDShouldBe(tableName, -id, newData)
}

func (ctx *TestContext) TableAtIDShouldBe(tableName string, id int64, data *gherkin.DataTable) error { // nolint
	// For that, we build a SQL request with only the attribute we are interested about (those
	// for the test data table) and we convert them to string (in SQL) to compare to table value.
	// Expect 'null' string in the table to check for nullness

	db := ctx.db()
	defer func() { /* #nosec */ _ = db.Close() }()

	var selects []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		selects = append(selects, cell.Value)
	}

	// define 'where' condition if needed
	where := ""
	if id != noID {
		if id < 0 {
			where = fmt.Sprintf(" WHERE ID <> %d ", -id)
		} else {
			where = fmt.Sprintf(" WHERE ID = %d ", id)
		}
	}

	// exec sql
	query := fmt.Sprintf("SELECT %s FROM `%s` %s", strings.Join(selects, ", "), tableName, where) // nolint: gosec
	sqlRows, err := db.Query(query)
	if err != nil {
		return err
	}
	dataCols := data.Rows[0].Cells
	iDataRow := 1
	sqlCols, _ := sqlRows.Columns() // nolint: gosec
	for sqlRows.Next() {
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

		nullValue := "null"
		pNullValue := &nullValue
		// checking that all columns of the test data table match the SQL row
		for iCol, dataCell := range data.Rows[iDataRow].Cells {
			if dataCell == nil {
				continue
			}
			colName := dataCols[iCol].Value
			dataValue := dataCell.Value
			sqlValue := rowValPtr[iCol].(**string)

			if *sqlValue == nil {
				sqlValue = &pNullValue
			}

			if (dataValue == "true" && **sqlValue == "1") || (dataValue == "false" && **sqlValue == "0") {
				continue
			}

			if dataValue != **sqlValue {
				return fmt.Errorf("not matching expected value at row %d, col %s, expected '%s', got: '%v'", iDataRow-1, colName, dataValue, **sqlValue)
			}
		}

		iDataRow++
	}

	// check that no row in the test data table has not been uncheck (if less rows in SQL result)
	if iDataRow < len(data.Rows) {
		return fmt.Errorf("there are less rows in the SQL results than expected. SQL: %d, expected: %d", iDataRow-1, len(data.Rows)-1)
	}
	return nil
}
