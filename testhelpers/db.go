//go:build !prod && !unit

package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"gopkg.in/yaml.v2"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

const (
	fixtureDir = "testdata" // special directory which is not included in binaries by the compiler
	utf8mb4    = "utf8mb4"
)

func init() { //nolint:gochecknoinits
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		appenv.ForceTestEnv()
	}
}

// SetupDBWithFixture creates a new DB connection, empties the DB, and loads a fixture.
// Note that the context must have a logger (set by logging.ContextWithLogger) and a config
// (set by testhelpers.CreateTestContext or testhelpers.CreateTestContextWithLogger),
// otherwise SetupDBWithFixture will panic.
func SetupDBWithFixture(ctx context.Context, fixtureNames ...string) *database.DB {
	appenv.ForceTestEnv()

	rawDB := OpenRawDBConnection(ctx)

	// Seed the DB
	emptyDBFromContextConfig(ctx, rawDB)
	for _, fixtureName := range fixtureNames {
		LoadFixture(ctx, rawDB, fixtureName)
	}

	// Return a new db connection
	db, err := database.Open(ctx, rawDB)
	if err != nil {
		panic(err)
	}

	return db
}

// SetupDBWithFixtureString creates a new DB connection, empties the DB,
// and loads fixtures from the strings (yaml with a tableName->[]dataRow map).
// Note that the context must have a logger (set by logging.ContextWithLogger) and a config
// (set by testhelpers.CreateTestContext or testhelpers.CreateTestContextWithLogger),
// otherwise SetupDBWithFixtureString will panic.
func SetupDBWithFixtureString(ctx context.Context, fixtures ...string) *database.DB {
	appenv.ForceTestEnv()

	rawDB := OpenRawDBConnection(ctx)

	// Seed the DB
	emptyDBFromContextConfig(ctx, rawDB)
	for _, fixture := range fixtures {
		loadFixtureChainFromString(ctx, rawDB, fixture)
	}

	// Return a new db connection
	db, err := database.Open(ctx, rawDB)
	if err != nil {
		panic(err)
	}

	return db
}

// OpenRawDBConnection creates a new connection to the DB specified in the config.
// Note that the context must have a logger (set by logging.ContextWithLogger) and a config
// (set by testhelpers.CreateTestContext or testhelpers.CreateTestContextWithLogger),
// otherwise OpenRawDBConnection will panic.
func OpenRawDBConnection(ctx context.Context) *sql.DB {
	appenv.ForceTestEnv()

	// needs actual config for connection to DB
	config := GetConfigFromContext(ctx)
	dbConfig, _ := app.DBConfig(config)
	if dbConfig.Params == nil {
		dbConfig.Params = make(map[string]string, 1)
	}
	dbConfig.Params["charset"] = utf8mb4
	logger := logging.LoggerFromContext(ctx)
	rawDB, err := database.OpenRawDBConnection(dbConfig.FormatDSN(), logger.IsRawSQLQueriesLoggingEnabled())
	if err != nil {
		panic(err)
	}
	return rawDB
}

// LoadFixture loads fixtures from `<current_pkg_dir>/testdata/<fileName>/` directory
// or `<current_pkg_dir>/testdata/<fileName>` file.
// Each file in this directory must be in yaml format.
// If a file name satisfies '*.chain.yaml' mask, the file is treated as a tableName->[]dataRow map.
// Otherwise, data will be loaded into table with the same name as the filename (without extension).
// Note that you should probably empty the DB before using this function.
// Note that the context must have a logger (set by logging.ContextWithLogger),
// otherwise LoadFixture will panic on logging.
func LoadFixture(ctx context.Context, db *sql.DB, fileName string) {
	appenv.ForceTestEnv()

	var filenames []string
	var err error
	filePath := filepath.Join(fixtureDir, fileName)
	stat, err := os.Stat(filePath)
	if err == nil && stat.IsDir() {
		files, err := os.ReadDir(filePath)
		if err != nil {
			panic(fmt.Errorf("unable to load fixture dir: %s", err.Error()))
		}
		filenames = make([]string, 0, len(files))
		for _, f := range files {
			filenames = append(filenames, f.Name())
		}
	} else {
		file, err := os.Lstat(filePath)
		if err != nil {
			panic(fmt.Errorf("unable to load fixture file: %s", err.Error()))
		}
		filePath = filepath.Dir(filePath)
		filenames = []string{file.Name()}
	}
	for _, filename := range filenames {
		var err error
		var data []byte
		data, err = os.ReadFile(filepath.Join(filePath, filename)) //nolint: gosec
		if err != nil {
			panic(err)
		}
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		if strings.HasSuffix(name, ".chain") {
			loadFixtureChainFromString(ctx, db, string(data))
		} else {
			var content []map[string]interface{}
			err = yaml.Unmarshal(data, &content)
			if err != nil {
				panic(err)
			}
			tableName := name
			logging.EntryFromContext(ctx).
				Infof("Loading data into %q:\n%s", tableName, string(data))
			InsertBatch(ctx, db, tableName, content)
		}
	}
}

func loadFixtureChainFromString(ctx context.Context, db *sql.DB, fixture string) {
	appenv.ForceTestEnv()

	var content yaml.MapSlice
	fixture = dedent.Dedent(fixture)
	fixture = strings.TrimSpace(strings.Replace(fixture, "\t", "  ", -1))
	bytesFixture := []byte(fixture)
	logging.EntryFromContext(ctx).Infof("Loading data chain:\n%s", bytesFixture)
	err := yaml.Unmarshal(bytesFixture, &content)
	if err != nil {
		panic(err)
	}
	for _, item := range content {
		data := make([]map[string]interface{}, 0, len(item.Value.([]interface{})))
		for _, row := range item.Value.([]interface{}) {
			rowSlice := row.(yaml.MapSlice)
			rowData := make(map[string]interface{}, len(rowSlice))
			for _, column := range rowSlice {
				rowData[column.Key.(string)] = column.Value
			}
			data = append(data, rowData)
		}
		InsertBatch(ctx, db, item.Key.(string), data)
	}
}

// InsertBatch insert the data into the table with the name given.
// Note that the context must have a logger (set by logging.ContextWithLogger),
// otherwise InsertBatch will panic on logging.
func InsertBatch(ctx context.Context, db *sql.DB, tableName string, data []map[string]interface{}) {
	appenv.ForceTestEnv()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0")
	if err != nil {
		panic(err)
	}

	for _, row := range data {
		var attributes []string
		var valueMarks []string
		var values []interface{}
		for k, v := range row {
			attributes = append(attributes, database.QuoteName(k))
			valueMarks = append(valueMarks, "?")
			values = append(values, v)
		}
		//nolint:gosec
		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
			tableName, strings.Join(attributes, ", "), strings.Join(valueMarks, ", "))
		_, err = tx.Exec(query, values...)
		if err != nil {
			panic(err)
		}
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

//nolint:gosec
func emptyDB(ctx context.Context, db *sql.DB, dbName string) error {
	appenv.ForceTestEnv()

	rows, err := db.QueryContext(ctx, `SELECT CONCAT(table_schema, '.', table_name)
                         FROM   information_schema.tables
                         WHERE  table_type   = 'BASE TABLE'
                           AND  table_schema = '`+dbName+`'
                           AND  table_name  != 'gorp_migrations'
                           AND  table_name  != 'user_batches'
                         ORDER BY table_name`)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()

		if rows.Err() != nil {
			panic(rows.Err())
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=0")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	for rows.Next() {
		var tableName string
		if scanErr := rows.Scan(&tableName); scanErr != nil {
			_, _ = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
			_ = tx.Rollback()
			return scanErr
		}
		// DELETE is MUCH faster than TRUNCATE on empty tables
		_, err = tx.Exec("DELETE FROM " + tableName)
		if err != nil {
			_, _ = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
			_ = tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS=1")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func emptyDBFromContextConfig(ctx context.Context, db *sql.DB) {
	config := GetConfigFromContext(ctx)
	dbConfig, err := app.DBConfig(config)
	mustNotBeError(err)
	mustNotBeError(emptyDB(ctx, db, dbConfig.DBName))
}
