package app_bdd_tests

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/config"
)

type testContext struct {
	application      *app.Application
	lastResponse     *http.Response
	lastResponseBody string
}

func (ctx *testContext) setupTestContext(interface{}) {

	config.Path = "../../conf/default.yaml"
	var err error
	ctx.application, err = app.New()
	if err != nil {
		fmt.Println("Unable to load app")
		panic(err)
	}

	err = ctx.emptyDB()
	if err != nil {
		fmt.Println("Unable to empty db")
		panic(err)
	}
}

func testRequest(ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string, error) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	return resp, string(respBody), nil
}

func (ctx *testContext) emptyDB() error { // FIXME, get the db name from config

	db := ctx.application.Database
	rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                                FROM   information_schema.tables
                                WHERE  table_type   = 'BASE TABLE'
                                  AND  table_schema = 'algorea_db'
                                  AND  table_name  != 'gorp_migrations'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			return err
		}
		if _, err := db.Exec("TRUNCATE TABLE " + tableName); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *testContext) iSendrequestGeneric(method string, path string, reqBody string) error {
	testServer := httptest.NewServer(ctx.application.HTTPHandler)
	defer testServer.Close()

	response, body, err := testRequest(testServer, method, path, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	ctx.lastResponse = response
	ctx.lastResponseBody = body

	return nil
}

/** Steps **/

func (ctx *testContext) dbHasTable(tableName string, data *gherkin.DataTable) error {

	db := ctx.application.Database
	var fields []string
	var marks []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		fields = append(fields, cell.Value)
		marks = append(marks, "?")
	}
	stmt, err := db.Prepare("INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES(" + strings.Join(marks, ", ") + ")")
	if err != nil {
		return err
	}
	for i := 1; i < len(data.Rows); i++ {
		var vals []interface{}
		for _, cell := range data.Rows[i].Cells {
			vals = append(vals, cell.Value)
		}
		if _, err = stmt.Exec(vals...); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *testContext) iSendrequestToWithBody(method string, path string, body *gherkin.DocString) error {
	return ctx.iSendrequestGeneric(method, path, body.Content)
}

func (ctx *testContext) iSendrequestTo(method string, path string) error {
	return ctx.iSendrequestGeneric(method, path, "")
}

func (ctx *testContext) itShouldBeAJSONArrayWithEntries(count int) error {
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("Unable to decode the response as JSON: %s\nData:%v", err, ctx.lastResponseBody)
	}

	if count != len(objmap) {
		return fmt.Errorf("The result does not have the expected length. Expected: %d, received: %d", count, len(objmap))
	}

	return nil
}

func (ctx *testContext) theResponseCodeShouldBe(code int) error {
	if code != ctx.lastResponse.StatusCode {
		return fmt.Errorf("expected response code to be: %d, but actual is: %d", code, ctx.lastResponse.StatusCode)
	}
	return nil
}

func (ctx *testContext) theResponseShouldMatchJSON(body *gherkin.DocString) (err error) {
	var expected, actual []byte
	var exp, act interface{}

	// re-encode expected response
	if err = json.Unmarshal([]byte(body.Content), &exp); err != nil {
		return
	}
	if expected, err = json.MarshalIndent(exp, "", "  "); err != nil {
		return
	}

	// re-encode actual response too
	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &act); err != nil {
		return fmt.Errorf("Unable to decode the response as JSON: %s\nData:%v", err, ctx.lastResponseBody)
	}
	if actual, err = json.MarshalIndent(act, "", "  "); err != nil {
		return
	}

	// the matching may be adapted per different requirements.
	if len(actual) != len(expected) {
		return fmt.Errorf(
			"expected json length: %d does not match actual: %d:\n%s",
			len(expected),
			len(actual),
			string(actual),
		)
	}

	for i, b := range actual {
		if b != expected[i] {
			return fmt.Errorf(
				"expected JSON does not match actual, showing up to last matched character:\n%s",
				string(actual[:i+1]),
			)
		}
	}
	return
}

func (ctx *testContext) tableShouldBe(tableName string, data *gherkin.DataTable) error {
	// For that, we build a SQL request with only the attribute we are interested about (those
	// for the test data table) and we convert them to string (in SQL) to compare to table value.
	// Expect 'null' string in the table to check for nullness

	db := ctx.application.Database
	var selects []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		selects = append(selects, fmt.Sprintf("CAST(IFNULL(%s,'NULL') as CHAR(50)) AS %s", cell.Value, cell.Value))
	}

	sqlRows, err := db.Query("SELECT " + strings.Join(selects, ", ") + " FROM " + tableName)
	if err != nil {
		return err
	}
	dataCols := data.Rows[0].Cells
	iDataRow := 1
	sqlCols, _ := sqlRows.Columns()
	for sqlRows.Next() {
		if iDataRow >= len(data.Rows) {
			return fmt.Errorf("There are more rows in the SQL results than expected. expected: %d", len(data.Rows)-1)
		}
		// Create a slice of string to represent each attribute value,
		// and a second slice to contain pointers to each item.
		rowValues := make([]string, len(sqlCols))
		rowValPtr := make([]interface{}, len(sqlCols))
		for i := range rowValues {
			rowValPtr[i] = &rowValues[i]
		}
		// Scan the result into the column pointers...
		if err := sqlRows.Scan(rowValPtr...); err != nil {
			return err
		}
		// checking that all columns of the test data table match the SQL row
		for iCol, dataCell := range data.Rows[iDataRow].Cells {
			colName := dataCols[iCol].Value
			dataValue := dataCell.Value
			sqlValue := rowValPtr[iCol].(*string)
			if dataValue != *sqlValue {
				return fmt.Errorf("Not matching expected value at row %d, col %s, expected '%s', got: '%v'", iDataRow-1, colName, dataValue, sqlValue)
			}
		}

		iDataRow++
	}

	// check that no row in teh test data table has not been uncheck (if less rows in SQL result)
	if iDataRow < len(data.Rows) {
		return fmt.Errorf("There are less rows in the SQL results than expected. SQL: %d, expected: %d", iDataRow-1, len(data.Rows)-1)
	}
	return nil
}

// FeatureContext binds the supported steps to the verifying functions
func FeatureContext(s *godog.Suite) {
	ctx := &testContext{}
	s.BeforeScenario(ctx.setupTestContext)

	s.Step(`^the database has the following table \'([\w\-_]*)\':$`, ctx.dbHasTable)

	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, ctx.iSendrequestTo)
	s.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with the following body:$`, ctx.iSendrequestToWithBody)
	s.Step(`^the response code should be (\d+)$`, ctx.theResponseCodeShouldBe)
	s.Step(`^the response should match json:$`, ctx.theResponseShouldMatchJSON)
	s.Step(`^it should be a JSON array with (\d+) entr(ies|y)$`, ctx.itShouldBeAJSONArrayWithEntries)
	s.Step(`^the table "([^"]*)" should be:$`, ctx.tableShouldBe)
}
