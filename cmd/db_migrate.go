package cmd

import (
	"fmt"
	"os"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

func init() {

	var migrateCmd = &cobra.Command{
		Use:   "db migrate",
		Short: "apply schema-change migrations to the database",
		Long:  `migrate uses go-pg migration tool under the hood supporting the same commands and an additional reset command`,
		Run: func(cmd *cobra.Command, args []string) {

			// load config
			if err := app.Config.Load(); err != nil {
				fmt.Println("Unable to load config: ", err)
				os.Exit(1)
			}

			// open DB
			migrations := &migrate.FileMigrationSource{Dir: "db/migrations"}
			db, err := database.DBConn(app.Config.Database)
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			// migrate
			n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
			if err != nil {
				fmt.Println("Unable to apply migration:", err)
			} else if n == 0 {
				fmt.Println("No migrations to apply!")
			} else {
				fmt.Printf("%d migration(s) applied successfully!\n", n)
			}

		},
	}

	rootCmd.AddCommand(migrateCmd)
}
