package cmd

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
)

//nolint:gosec
func init() { //nolint:gochecknoinits
	restoreCmd := &cobra.Command{
		Use:   "db-restore [environment]",
		Short: "load the last db schema",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()
			if appenv.IsEnvProd() {
				fmt.Println("'db-restore' must not be run in 'prod' env!")
				os.Exit(1)
			}

			// load config
			dbConf, err := app.DBConfig(app.LoadConfig())
			if err != nil {
				fmt.Println("Unable to load the database config: ", err)
				os.Exit(1)
			}

			// open DB
			var db *sql.DB
			db, err = sql.Open("mysql", dbConf.FormatDSN())
			assertNoError(err, "Unable to connect to the database: ")
			defer func() { _ = db.Close() }()

			tx, err := db.Begin()
			assertNoError(err, "Unable to start a transaction: ")

			_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
			assertNoError(err, "Unable to query the database: ")

			// remove all tables from DB
			var rows *sql.Rows
			rows, err = db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                            FROM   information_schema.tables
                            WHERE  table_type   = 'BASE TABLE'
                              AND  table_schema = '` + dbConf.DBName + "'")
			assertNoError(err, "Unable to query the database: ")

			defer func() {
				_ = rows.Close()

				if rows.Err() != nil {
					panic(rows.Err())
				}
			}()

			for rows.Next() {
				var tableName string
				assertNoError(rows.Scan(&tableName), "Unable to parse the database result: ")
				_, err = tx.Exec("DROP TABLE " + tableName)
				assertNoError(err, "Unable to drop table: ")
			}

			_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
			assertNoError(err, "Unable to query the database: ")

			err = tx.Commit()
			assertNoError(err, "Unable to commit the transaction: ")

			// restore the schema
			// note: current solution is not really great as it makes some assumptions of the config :-/
			host, port, err := net.SplitHostPort(dbConf.Addr)
			if err != nil {
				host = dbConf.Addr
				port = "3306"
			}
			command := exec.Command(
				"mysql",
				"-h"+host,
				"-P"+port,
				"-D"+dbConf.DBName,
				"-u"+dbConf.User,
				"-p"+dbConf.Passwd,
				"--protocol=TCP",
				"-e"+"source db/schema/schema.sql",
			)
			fmt.Println("mysql importing dump...")
			var output []byte
			output, err = command.CombinedOutput()
			if err != nil {
				fmt.Printf("Command finished with error: %v\n", err)
				fmt.Printf("Output:\n%s", output)
				os.Exit(1)
			}

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(restoreCmd)
}

func assertNoError(err error, message string) {
	if err != nil {
		fmt.Println(message, err)
		os.Exit(1)
	}
}
