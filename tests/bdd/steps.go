package app_bdd_tests

import (
	"database/sql"
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
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

type testContext struct {
	application      http.Handler
	db               *sql.DB
	lastResponse     *http.Response
	lastResponseBody string
}

func (ctx *testContext) setupTestContext(interface{}) {

	app.ConfigFile = "../../conf/default.yaml"
	if err := app.Config.Load(); err != nil {
		fmt.Println("Unable to load config")
		panic(err)
	}
	ctx.application, _ = app.New()

	db, err := database.DBConn(app.Config.Database)
	if err != nil {
		fmt.Println("Unable to load db")
		panic(err)
	}
	ctx.db = db

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

	rows, err := ctx.db.Query(`SELECT CONCAT(table_schema, '.', table_name)
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
		if _, err := ctx.db.Exec("TRUNCATE TABLE " + tableName); err != nil {
			return err
		}
	}
	return nil
}

/** Steps **/

func (ctx *testContext) dbSeed(tableName string, data *gherkin.DataTable) error {

	var fields []string
	var marks []string
	head := data.Rows[0].Cells
	for _, cell := range head {
		fields = append(fields, cell.Value)
		marks = append(marks, "?")
	}
	stmt, err := ctx.db.Prepare("INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES(" + strings.Join(marks, ", ") + ")")
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

func (ctx *testContext) makeRequest(method string, path string) error {
	testServer := httptest.NewServer(ctx.application)
	defer testServer.Close()

	response, body, err := testRequest(testServer, method, path, nil)
	if err != nil {
		return err
	}
	ctx.lastResponse = response
	ctx.lastResponseBody = body

	return nil
}

func (ctx *testContext) itShouldBeAJSONArrayWithEntries(count int) error {
	var objmap []map[string]*json.RawMessage

	if err := json.Unmarshal([]byte(ctx.lastResponseBody), &objmap); err != nil {
		return fmt.Errorf("Unable to decode the response as JSON: %s", err)
	}

	if count != len(objmap) {
		return fmt.Errorf("The result does not have the expected length. Expected: %d, received: %d", count, len(objmap))
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	ctx := &testContext{}
	s.BeforeScenario(ctx.setupTestContext)

	s.Step(`^the database has the following table \'([\w\-_]*)\':$`, ctx.dbSeed)
	s.Step(`^I make a (GET) (/[\w\/]*)$`, ctx.makeRequest)
	s.Step(`^it should be a JSON array with (\d+) entr(ies|y)$`, ctx.itShouldBeAJSONArrayWithEntries)
}
