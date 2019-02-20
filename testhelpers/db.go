package testhelpers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

const fixtureDir = "testdata" // special directory which is not included in binaries by the compile

// LoadFixture load the fixtures from `<current_pkg_dir>/testdata/<dirname/`.
// Each file in this directory mush be in yaml format and will be loaded into table
//  with the same name as the filename (without extension)
// Note that you should probably empty the DB before using this function.
func LoadFixture(t *testing.T, db *sql.DB, dirName string) {
	dirPath := filepath.Join(fixtureDir, dirName)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		t.Fatalf("Unable to load fixture dir: %s", err.Error())
	}
	for _, f := range files {
		var err error
		var data []byte
		filename := f.Name()
		tableName := strings.TrimSuffix(filename, filepath.Ext(filename))
		data, err = ioutil.ReadFile(filepath.Join(fixtureDir, dirName, filename)) // nolint: gosec
		if err != nil {
			t.Fatal(err)
		}
		content := make([]map[string]interface{}, 0)
		err = yaml.Unmarshal(data, &content)
		if err != nil {
			t.Fatal(err)
		}
		InsertBatch(t, db, tableName, content)
	}
}

// InsertBatch insert the data into the table with the name given
func InsertBatch(t *testing.T, db *sql.DB, tableName string, data []map[string]interface{}) {
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
			t.Fatal(err)
		}
	}

}

// EmptyDB empties all tables of the give database
// nolint: gosec
func EmptyDB(t *testing.T, db *sql.DB, dbName string) {

	rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                         FROM   information_schema.tables
                         WHERE  table_type   = 'BASE TABLE'
                           AND  table_schema = '` + dbName + `'
                           AND  table_name  != 'gorp_migrations'`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec("TRUNCATE TABLE " + tableName)
		if err != nil {
			t.Fatal(err)
		}
	}
}
