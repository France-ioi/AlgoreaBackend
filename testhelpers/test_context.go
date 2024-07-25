//go:build !prod

package testhelpers

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"

	"bou.ke/monkey"
	"github.com/CloudyKit/jet"
	"github.com/cucumber/godog"
	_ "github.com/go-sql-driver/mysql"      // use to force database/sql to use mysql
	"github.com/sirupsen/logrus/hooks/test" //nolint:depguard
	"github.com/thingful/httpmock"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/loggingtest"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
)

type dbquery struct {
	sql    string
	values []interface{}
}

// TestContext implements context for tests.
type TestContext struct {
	application          *app.Application // do NOT call it directly, use `app()`
	userID               int64            // userID that will be used for the next requests
	user                 string           // user reference of the logged user
	featureQueries       []dbquery
	lastResponse         *http.Response
	lastResponseBody     string
	logsHook             *loggingtest.Hook
	logsRestoreFunc      func()
	inScenario           bool
	db                   *sql.DB
	dbTableData          map[string]*godog.Table
	templateSet          *jet.Set
	requestHeaders       map[string][]string
	referenceToIDMap     map[string]int64
	dbTables             map[string]map[string]map[string]interface{}
	currentThreadKey     string
	allUsersGroup        string
	needPopulateDatabase bool
}

const (
	testAccessToken = "testsessiontestsessiontestsessio"
	testSessionID   = 123451234512345
)

// SetupTestContext initializes the test context. Called before each scenario.
func (ctx *TestContext) SetupTestContext(sc *godog.Scenario) {
	var logHook *test.Hook
	logHook, ctx.logsRestoreFunc = log.MockSharedLoggerHook()
	ctx.logsHook = &loggingtest.Hook{Hook: logHook}

	ctx.setupApp()
	ctx.userID = 0 // not set
	ctx.lastResponse = nil
	ctx.lastResponseBody = ""
	ctx.inScenario = true
	ctx.requestHeaders = map[string][]string{}
	ctx.db = ctx.openDB()
	ctx.dbTableData = make(map[string]*godog.Table)
	ctx.templateSet = ctx.constructTemplateSet()
	ctx.dbTables = make(map[string]map[string]map[string]interface{})
	ctx.needPopulateDatabase = false

	ctx.initReferences(sc)

	// reset the seed to get predictable results on PRNG for tests
	rand.Seed(1)

	err := ctx.initDB()
	if err != nil {
		fmt.Println("Unable to empty db")
		panic(err)
	}
}

// initReferences initializes the referenceToIDMap
// generating unique IDs for references. The generated IDs have the same
// sorting order as the references.
func (ctx *TestContext) initReferences(sc *godog.Scenario) {
	collectedReferences := collectReferences(sc)
	ctx.referenceToIDMap = make(map[string]int64, len(collectedReferences))
	for index, reference := range collectedReferences {
		id := int64(1000000000000000000) + int64(index)
		ctx.referenceToIDMap[reference] = id
	}
}

func collectReferencesInText(text string, referencesMap map[string]struct{}) {
	for _, match := range referenceRegexp.FindAllString(text, -1) {
		if match[0] != ReferencePrefix {
			match = match[1:]
		}
		referencesMap[match] = struct{}{}
	}
}

// collectReferences collects all references in a scenario and returns them sorted.
func collectReferences(sc *godog.Scenario) []string {
	referencesMap := make(map[string]struct{})

	for _, step := range sc.Steps {
		collectReferencesInText(step.Text, referencesMap)
		if step.Argument == nil {
			continue
		}
		if table := step.Argument.DataTable; table != nil {
			for _, row := range table.Rows {
				for _, cell := range row.Cells {
					collectReferencesInText(cell.Value, referencesMap)
				}
			}
		}
		if docString := step.Argument.DocString; docString != nil {
			collectReferencesInText(docString.Content, referencesMap)
		}
	}

	// Add globally defined references
	referencesMap["@AllUsers"] = struct{}{}

	references := make([]string, 0, len(referencesMap))
	for reference := range referencesMap {
		references = append(references, reference)
	}
	sort.Strings(references)

	return references
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

// ScenarioTeardown is called after each scenario to remove stubs.
func (ctx *TestContext) ScenarioTeardown(*godog.Scenario, error) {
	RestoreDBTime()
	monkey.UnpatchAll()
	ctx.logsRestoreFunc()

	defer func() {
		if err := httpmock.AllStubsCalled(); err != nil {
			panic(err) // godog doesn't allow to return errors from handlers (see https://github.com/cucumber/godog/issues/88)
		}
		httpmock.DeactivateAndReset()
	}()

	ctx.tearDownApp()
}

func testRequest(ts *httptest.Server, method, path string, headers map[string][]string, body io.Reader) (*http.Response, string, error) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		return nil, "", err
	}

	// add headers
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	// execute the query
	resp, err := client.Do(req)
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

// openDB opens a connection to the database.
// We use instrumented-mysql driver to log all queries.
func (ctx *TestContext) openDB() *sql.DB {
	if ctx.db == nil {
		var err error
		config, _ := app.DBConfig(ctx.application.Config)
		ctx.db, err = sql.Open("instrumented-mysql", config.FormatDSN())
		if err != nil {
			fmt.Println("Unable to connect to the database: ", err)
			os.Exit(1)
		}
	}

	return ctx.db
}

func (ctx *TestContext) emptyDB() error {
	config, _ := app.DBConfig(ctx.application.Config)
	return emptyDB(ctx.db, config.DBName)
}

func (ctx *TestContext) initDB() error {
	err := ctx.emptyDB()
	if err != nil {
		return err
	}

	if len(ctx.featureQueries) > 0 {
		tx, err := ctx.db.Begin()
		if err != nil {
			return err
		}
		_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0")
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		for _, query := range ctx.featureQueries {
			_, err = tx.Exec(query.sql, query.values)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
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
	}

	return nil
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}
