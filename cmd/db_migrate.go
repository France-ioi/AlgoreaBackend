package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func init() { // nolint:gochecknoinits

	var migrateCmd = &cobra.Command{
		Use:   "db-migrate [environment]",
		Short: "apply schema-change migrations to the database",
		Long:  `migrate uses go-pg migration tool under the hood supporting the same commands and an additional reset command`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()

			// open DB
			migrations := &migrate.FileMigrationSource{Dir: "db/migrations"}
			var db *sql.DB
			databaseConfig, err := app.DBConfig(app.LoadConfig())
			if err != nil {
				fmt.Println("Unable to load the database config: ", err)
				os.Exit(1)
			}
			databaseConfig.ParseTime = true
			db, err = sql.Open("mysql", databaseConfig.FormatDSN())
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			// migrate
			var n int
			n, err = migrate.Exec(db, "mysql", migrations, migrate.Up)
			switch {
			case err != nil:
				fmt.Println("Unable to apply migration:", err)
				os.Exit(1)
			case n == 0:
				fmt.Println("No migrations to apply!")
			default:
				fmt.Printf("%d migration(s) applied successfully!\n", n)

				var gormDB *database.DB
				gormDB, err = database.Open(db)
				assertNoError(err, "Cannot open GORM db connection: ")
				fmt.Print("Running ANALYZE TABLE attempts\n")
				_, err = db.Exec("ANALYZE TABLE `attempts`")
				assertNoError(err, "Cannot execute ANALYZE TABLE")
				fmt.Print("Running ANALYZE TABLE `groups`\n")
				_, err = db.Exec("ANALYZE TABLE `groups`")
				assertNoError(err, "Cannot execute ANALYZE TABLE")
				fmt.Print("Running ANALYZE TABLE `items`\n")
				_, err = db.Exec("ANALYZE TABLE `items`")
				assertNoError(err, "Cannot execute ANALYZE TABLE")
				err = database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
					fmt.Print("Running GroupGroupStore.After()\n")
					assertNoError(store.GroupGroups().After(), "Cannot compute groups_groups") // calls createNewAncestors()
					fmt.Print("Running ItemItemStore.After()\n")
					assertNoError(store.ItemItems().After(), "Cannot compute items_items") // calls createNewAncestors() & computeAllAccess()
					return nil
				})
			}

			if db.Close() != nil {
				fmt.Println("Cannot close DB connection:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(migrateCmd)
}
