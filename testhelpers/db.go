package testhelpers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

const fixtureDir = "testdata" // special directory which is not included in binaries by the compile

// SetupDBWithFixture creates a new DB connection, empties the DB, and loads a fixture
func SetupDBWithFixture(fixtureName string) *database.DB {
	rawDb, err := OpenRawDBConnection()

	// Seed the DB
	EmptyDB(rawDb)
	LoadFixture(rawDb, fixtureName)

	// Return a new db connection
	var db *database.DB
	db, err = database.Open(rawDb)
	if err != nil {
		panic(err)
	}

	return db
}

// SetupDBWithFixture creates a new connection to the DB specified in the config
func OpenRawDBConnection() (*sql.DB, error) {
	// needs actual config for connection to DB
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}
	var rawDb *sql.DB
	rawDb, err = database.OpenRawDBConnection(conf.Database.Connection.FormatDSN())
	if err != nil {
		panic(err)
	}
	return rawDb, err
}

// LoadFixture load the fixtures from `<current_pkg_dir>/testdata/<dirname/`.
// Each file in this directory mush be in yaml format and will be loaded into table
//  with the same name as the filename (without extension)
// Note that you should probably empty the DB before using this function.
func LoadFixture(db *sql.DB, dirName string) {
	dirPath := filepath.Join(fixtureDir, dirName)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		panic(fmt.Errorf("unable to load fixture dir: %s", err.Error()))
	}
	for _, f := range files {
		var err error
		var data []byte
		filename := f.Name()
		tableName := strings.TrimSuffix(filename, filepath.Ext(filename))
		data, err = ioutil.ReadFile(filepath.Join(fixtureDir, dirName, filename)) // nolint: gosec
		if err != nil {
			panic(err)
		}
		content := make([]map[string]interface{}, 0)
		err = yaml.Unmarshal(data, &content)
		if err != nil {
			panic(err)
		}
		InsertBatch(db, tableName, content)
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
		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(attributes, ", "), strings.Join(valueMarks, ", ")) // nolint: gosec
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
		_, err = db.Exec("TRUNCATE TABLE " + tableName)
		if err != nil {
			panic(err)
		}
	}
}

// EmptyDB empties all tables of the database specified in the config
func EmptyDB(db *sql.DB) {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	emptyDB(db, conf.Database.Connection.DBName)
}
