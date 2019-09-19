package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

const (
	migrateStatementBegin = "-- +migrate StatementBegin\n"
	migrateStatementEnd   = "-- +migrate StatementEnd\n"
)

// nolint:gosec
func init() { // nolint:gochecknoinits,gocyclo

	var dbGenMigrationsCmd = &cobra.Command{
		Use:   "db-gen-migrations  [environment]",
		Short: "generate migrations",
		Run: func(cmd *cobra.Command, args []string) {
			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnv("dev")

			// load config
			conf := config.Load()
			if appenv.IsEnvProd() {
				fmt.Println("'db-gen-migrations' must not be run in 'prod' env!")
				os.Exit(1)
			}

			if conf.Database.Connection.Net == "" {
				fmt.Println("database.connection.net should be set")
				os.Exit(1)
			}

			// open DB
			rawdb, err := sql.Open("mysql", conf.Database.Connection.FormatDSN())
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			commentRegexp := regexp.MustCompile(" COMMENT '(.+)'$")

			renamedColumns := map[string]string{}
			var renameMigrationUp, renameMigrationDown string

			var rows *sql.Rows
			tables := getTables(rawdb, conf.Database.Connection.DBName)
			for _, table := range tables {
				var renameStatements, renameBackStatements []string
				rows, err = rawdb.Query("SHOW CREATE TABLE " + table)
				if err != nil {
					panic(err)
				}
				for rows.Next() {
					var tableName, spec string
					if err = rows.Scan(&tableName, &spec); err != nil {
						panic(err)
					}
					specRows := strings.Split(spec, "\n")
					for _, specRow := range specRows {
						row := strings.TrimSpace(specRow)
						if row[0] != '`' {
							var indexName, convertedIndexName string
							if strings.HasPrefix(row, "UNIQUE KEY `") {
								indexName = strings.SplitN(strings.TrimPrefix(row, "UNIQUE KEY `"), "` ", 2)[0]
							}
							if strings.HasPrefix(row, "KEY `") {
								indexName = strings.SplitN(strings.TrimPrefix(row, "KEY `"), "` ", 2)[0]
							}
							convertedIndexName = convertColumnName(indexName)
							if convertedIndexName != indexName {
								renameStatements = append(renameStatements,
									"\tRENAME INDEX "+database.QuoteName(indexName)+" TO "+database.QuoteName(convertedIndexName))
								renameBackStatements = append(renameBackStatements,
									"\tRENAME INDEX "+database.QuoteName(convertedIndexName)+" TO "+database.QuoteName(indexName))
							}
							continue
						}
						columnName := strings.SplitN(row, "`", 3)[1]
						convertedName := convertColumnName(columnName)
						if convertedName != columnName {
							renamedColumns[columnName] = convertedName
							renameStatements = append(renameStatements, "\tRENAME COLUMN "+database.QuoteName(columnName)+
								" TO "+database.QuoteName(convertedName))
							renameBackStatements = append(renameBackStatements, "\tRENAME COLUMN "+database.QuoteName(convertedName)+
								" TO "+database.QuoteName(columnName))
						}
					}
				}
				_ = rows.Close()
				table = strings.SplitN(table, ".", 2)[1]
				if len(renameStatements) > 0 {
					renameMigrationUp += "ALTER TABLE " + database.QuoteName(table) + "\n" + strings.Join(renameStatements, ",\n") + ";\n"
					renameMigrationDown += "ALTER TABLE " + database.QuoteName(table) + "\n" + strings.Join(renameBackStatements, ",\n") + ";\n"
				}
			}

			var commentsMigrationUp, commentsMigrationDown string
			for _, table := range tables {
				var modifyStatements, modifyBackStatements []string
				rows, err = rawdb.Query("SHOW CREATE TABLE " + table)
				if err != nil {
					panic(err)
				}
				for rows.Next() {
					var tableName, spec string
					if err = rows.Scan(&tableName, &spec); err != nil {
						panic(err)
					}
					specRows := strings.Split(spec, "\n")
					for _, specRow := range specRows {
						row := strings.TrimSpace(specRow)
						row = strings.TrimRight(row, ",")
						if row[0] != '`' {
							continue
						}
						commentMatch := commentRegexp.FindStringSubmatch(row)
						if len(commentMatch) > 0 {
							comment := commentMatch[1]
							comment = fullTextReplace(renamedColumns, comment)
							if comment != commentMatch[1] {
								newSpec := commentRegexp.ReplaceAllString(row, " COMMENT '"+comment+"'")
								modifyStatements = append(modifyStatements, "\tMODIFY "+newSpec)
								modifyBackStatements = append(modifyBackStatements, "\tMODIFY "+row)
							}
						}
					}
				}
				_ = rows.Close()
				table = strings.SplitN(table, ".", 2)[1]
				if len(modifyStatements) > 0 {
					commentsMigrationUp += "ALTER TABLE " + database.QuoteName(table) + "\n" + strings.Join(modifyStatements, ",\n") + ";\n"
					commentsMigrationDown += "ALTER TABLE " + database.QuoteName(table) + "\n" + strings.Join(modifyBackStatements, ",\n") + ";\n"
				}
			}

			var triggersMigrationUp, triggersMigrationDown string
			rows, err = rawdb.Query("SHOW TRIGGERS")
			if err != nil {
				panic(err)
			}
			for rows.Next() {
				var trigger, event, table, statement, timing, created, sqlMode, definer, characterSetClient,
					collationConnection, databaseCollation string
				if err = rows.Scan(&trigger, &event, &table, &statement, &timing, &created, &sqlMode, &definer, &characterSetClient,
					&collationConnection, &databaseCollation); err != nil {
					panic(err)
				}
				modifiedStatement := fullTextReplace(renamedColumns, statement)
				if modifiedStatement != statement {
					triggerRows, err := rawdb.Query("SHOW CREATE TRIGGER " + database.QuoteName(trigger))
					if err != nil {
						panic(err)
					}
					for triggerRows.Next() {
						if err = triggerRows.Scan(&trigger, &sqlMode, &statement, &characterSetClient, &collationConnection,
							&databaseCollation, &created); err != nil {
							panic(err)
						}
						modifiedStatement = fullTextReplace(renamedColumns, statement)
						triggersMigrationUp += "DROP TRIGGER " + database.QuoteName(trigger) + ";\n"
						triggersMigrationUp += migrateStatementBegin
						triggersMigrationUp += modifiedStatement + "\n"
						triggersMigrationUp += migrateStatementEnd
						triggersMigrationDown += "DROP TRIGGER " + database.QuoteName(trigger) + ";\n"
						triggersMigrationDown += migrateStatementBegin
						triggersMigrationDown += statement + "\n"
						triggersMigrationDown += migrateStatementEnd
					}
					_ = triggerRows.Close()
				}
			}
			_ = rows.Close()

			fmt.Printf("\n\n\n/* Renamed columns:\n %#v\n*/\n\n", renamedColumns)

			result := "-- +migrate Up\n" + commentsMigrationUp + "\n\n" + renameMigrationUp + "\n\n" + triggersMigrationUp + "\n\n" +
				"-- +migrate Down\n" + renameMigrationDown + "\n\n" + commentsMigrationDown + "\n\n" + triggersMigrationDown

			fmt.Println(result)
		},
	}

	rootCmd.AddCommand(dbGenMigrationsCmd)
}

func fullTextReplace(renamedColumns map[string]string, text string) string {
	for oldName, newName := range renamedColumns {
		columnRegexp := regexp.MustCompile("\\b" + regexp.QuoteMeta(oldName) + "\\b")
		text = columnRegexp.ReplaceAllString(text, newName)
	}
	return text
}

// nolint:gosec
func getTables(db *sql.DB, dbName string) []string {
	rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                         FROM   information_schema.tables
                         WHERE  table_type   = 'BASE TABLE'
                           AND  table_schema = '` + dbName + `'
                           AND  table_name  != 'gorp_migrations'`)
	if err != nil {
		panic(err)
	}
	defer func() { _ = rows.Close() }()

	var result []string
	for rows.Next() {
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			panic(err)
		}
		result = append(result, tableName)
	}
	return result
}

func convertColumnName(columnName string) string {
	if columnName == "sPasswordMd5" {
		return "password_md5"
	}

	snakeCaseName := toSnakeCase(columnName)
	underscoreIndex := strings.IndexByte(snakeCaseName, '_')
	prefix := ""
	if underscoreIndex > 0 {
		prefix = snakeCaseName[:underscoreIndex]
	}

	if prefix == "id" {
		snakeCaseName = snakeCaseName[3:] + "_id"
	}

	if map[string]bool{"nb": true, "b": true, "i": true, "s": true}[prefix] {
		snakeCaseName = snakeCaseName[underscoreIndex+1:]
	}

	return snakeCaseName
}

// toSnakeCase converts the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func toSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) &&
			((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
