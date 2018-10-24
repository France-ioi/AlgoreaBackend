package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/spf13/cobra"
)

func init() {

	var restoreCmd = &cobra.Command{
		Use:   "db-restore",
		Short: "load the last db schema",
		Run: func(cmd *cobra.Command, args []string) {

			// load config
			if err := app.Config.Load(); err != nil {
				fmt.Println("Unable to load config: ", err)
				os.Exit(1)
			}

			// open DB
			db, err := database.DBConn(app.Config.Database)
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			// remove all tables from DB
			rows, err := db.Query(`SELECT CONCAT(table_schema, '.', table_name)
	                           FROM   information_schema.tables
														 WHERE  table_type   = 'BASE TABLE'
														   AND  table_schema = '` + app.Config.Database.DBName + "'")
			if err != nil {
				fmt.Println("Unable to query the database: ", err)
				os.Exit(1)
			}
			defer rows.Close()

			for rows.Next() {
				var tableName string
				if err = rows.Scan(&tableName); err != nil {
					fmt.Println("Unable to parse the database result: ", err)
					os.Exit(1)
				}
				if _, err := db.Exec("DROP TABLE " + tableName); err != nil {
					fmt.Println("Unable to drop table: ", err)
					os.Exit(1)
				}
			}

			// restore the schema
			// note: current solution is not really great as it makes some assumptions of the config :-/
			command := exec.Command(
				"mysql",
				"-h"+app.Config.Database.Addr,
				"-D"+app.Config.Database.DBName,
				"-u"+app.Config.Database.User,
				"-p"+app.Config.Database.Passwd,
				"--protocol=TCP",
				"-e"+"source db/schema/20181024.sql",
			)
			fmt.Println("mysql importing dump...")
			err = command.Run()
			if err != nil {
				fmt.Printf("Command finished with error: %v", err)
				os.Exit(1)
			}

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(restoreCmd)
}
