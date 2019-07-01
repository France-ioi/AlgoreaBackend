package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

// nolint:gosec
func init() { // nolint:gochecknoinits

	var restoreCmd = &cobra.Command{
		Use:   "db-restore <environment>",
		Short: "load the last db schema",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// load config
			conf := config.Load(args[0])

			// open DB
			var db *sql.DB
			db, err = sql.Open("mysql", conf.Database.Connection.FormatDSN())
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			// remove all tables from DB
			var rows *sql.Rows
			rows, err = db.Query(`SELECT CONCAT(table_schema, '.', table_name)
                            FROM   information_schema.tables
                            WHERE  table_type   = 'BASE TABLE'
                              AND  table_schema = '` + conf.Database.Connection.DBName + "'")
			if err != nil {
				fmt.Println("Unable to query the database: ", err)
				os.Exit(1)
			}

			defer func() {
				_ = rows.Close()
				_ = db.Close()
			}()

			for rows.Next() {
				var tableName string
				if err = rows.Scan(&tableName); err != nil { // nolint: vetshadow
					fmt.Println("Unable to parse the database result: ", err)
					os.Exit(1)
				}
				_, err = db.Query("DROP TABLE " + tableName)
				if err != nil {
					fmt.Println("Unable to drop table: ", err)
					os.Exit(1)
				}
			}

			// restore the schema
			// note: current solution is not really great as it makes some assumptions of the config :-/
			command := exec.Command(
				"mysql",
				"-h"+conf.Database.Connection.Addr,
				"-D"+conf.Database.Connection.DBName,
				"-u"+conf.Database.Connection.User,
				"-p"+conf.Database.Connection.Passwd,
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
