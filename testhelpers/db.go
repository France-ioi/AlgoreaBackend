package testhelpers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"gopkg.in/yaml.v2"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

const fixtureDir = "testdata" // special directory which is not included in binaries by the compile

func init() { // nolint:gochecknoinits
	conf := config.Load("test")
	// Apply the config to the global logger
	logging.SharedLogger.Configure(conf.Logging)
}

// SetupDBWithFixture creates a new DB connection, empties the DB, and loads a fixture
func SetupDBWithFixture(fixtureNames ...string) *database.DB {
	rawDb, err := OpenRawDBConnection()
	if err != nil {
		panic(err)
	}

	// Seed the DB
	EmptyDB(rawDb)
	for _, fixtureName := range fixtureNames {
		LoadFixture(rawDb, fixtureName)
	}

	// Return a new db connection
	var db *database.DB
	db, err = database.Open(rawDb)
	if err != nil {
		panic(err)
	}

	return db
}

// SetupDBWithFixtureString creates a new DB connection, empties the DB,
// and loads fixtures from the strings (yaml with a tableName->[]dataRow map)
func SetupDBWithFixtureString(fixtures ...string) *database.DB {
	rawDb, err := OpenRawDBConnection()
	if err != nil {
		panic(err)
	}

	// Seed the DB
	EmptyDB(rawDb)

	for _, fixture := range fixtures {
		loadFixtureChainFromString(rawDb, fixture)
	}

	// Return a new db connection
	var db *database.DB
	db, err = database.Open(rawDb)
	if err != nil {
		panic(err)
	}

	return db
}

// OpenRawDBConnection creates a new connection to the DB specified in the config
func OpenRawDBConnection() (*sql.DB, error) {
	// needs actual config for connection to DB
	conf := config.Load("test")
	rawDb, err := database.OpenRawDBConnection(conf.Database.Connection.FormatDSN())
	if err != nil {
		panic(err)
	}
	return rawDb, err
}

// LoadFixture loads fixtures from `<current_pkg_dir>/testdata/<fileName>/` directory
// or `<current_pkg_dir>/testdata/<fileName>` file.
// Each file in this directory must be in yaml format.
// If a file name satisfies '*.chain.yaml' mask, the file is treated as a tableName->[]dataRow map.
// Otherwise, data will be loaded into table with the same name as the filename (without extension).
// Note that you should probably empty the DB before using this function.
func LoadFixture(db *sql.DB, fileName string) {
	var files []os.FileInfo
	var err error
	filePath := filepath.Join(fixtureDir, fileName)
	stat, err := os.Stat(filePath)
	if err == nil && stat.IsDir() {
		files, err = ioutil.ReadDir(filePath)
		if err != nil {
			panic(fmt.Errorf("unable to load fixture dir: %s", err.Error()))
		}
	} else {
		file, err := os.Lstat(filePath)
		if err != nil {
			panic(fmt.Errorf("unable to load fixture file: %s", err.Error()))
		}
		filePath = filepath.Dir(filePath)
		files = []os.FileInfo{file}
	}
	for _, f := range files {
		var err error
		var data []byte
		filename := f.Name()
		data, err = ioutil.ReadFile(filepath.Join(filePath, filename)) // nolint: gosec
		if err != nil {
			panic(err)
		}
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		if strings.HasSuffix(name, ".chain") {
			loadFixtureChainFromString(db, string(data))
		} else {
			var content []map[string]interface{}
			err = yaml.Unmarshal(data, &content)
			if err != nil {
				panic(err)
			}
			tableName := name
			logging.SharedLogger.Infof("Loading data into %q:\n%s", tableName, string(data))
			InsertBatch(db, tableName, content)
		}
	}
}

func loadFixtureChainFromString(db *sql.DB, fixture string) {
	var content map[string][]map[string]interface{}
	fixture = dedent.Dedent(fixture)
	fixture = strings.TrimSpace(strings.Replace(fixture, "\t", "  ", -1))
	bytesFixture := []byte(fixture)
	logging.SharedLogger.Infof("Loading data chain:\n%s", bytesFixture)
	err := yaml.Unmarshal(bytesFixture, &content)
	if err != nil {
		panic(err)
	}
	for tableName, tableData := range content {
		InsertBatch(db, tableName, tableData)
	}
}

// InsertBatch insert the data into the table with the name given
func InsertBatch(db *sql.DB, tableName string, data []map[string]interface{}) {
	for _, row := range data {
		var attributes []string
		var valueMarks []string
		var values []interface{}
		for k, v := range row {
			attributes = append(attributes, k)
			valueMarks = append(valueMarks, "?")
			values = append(values, v)
		}
		// nolint:gosec
		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
			tableName, strings.Join(attributes, ", "), strings.Join(valueMarks, ", "))
		_, err := db.Exec(query, values...)
		if err != nil {
			panic(err)
		}
	}

}

// nolint: gosec
func emptyDB(db *sql.DB, dbName string) {

	rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                         FROM   information_schema.tables
                         WHERE  table_type   = 'BASE TABLE'
                           AND  table_schema = '` + dbName + `'
                           AND  table_name  != 'gorp_migrations'`)
	if err != nil {
		panic(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			panic(err)
		}
		// DELETE is MUCH faster than TRUNCATE on empty tables
		_, err = db.Exec("DELETE FROM " + tableName)
		if err != nil {
			panic(err)
		}
	}
}

// EmptyDB empties all tables of the database specified in the config
func EmptyDB(db *sql.DB) {
	conf := config.Load("test")
	emptyDB(db, conf.Database.Connection.DBName)
}
