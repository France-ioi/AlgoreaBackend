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

func init() { //nolint:gochecknoinits
	migrateCmd := &cobra.Command{
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
			var applied int
			for {
				var n int
				n, err = migrate.ExecMax(db, "mysql", migrations, migrate.Up, 1)
				if err != nil {
					fmt.Println("\nUnable to apply migration:", err)
					os.Exit(1)
				}
				applied += n
				if n == 0 {
					break
				}
				fmt.Print(".")
			}
			fmt.Print("\n")
			switch {
			case applied == 0:
				fmt.Println("No migrations to apply!")
			default:
				fmt.Printf("%d migration(s) applied successfully!\n", applied)

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
				assertNoError(recomputeDBCaches(gormDB), "Cannot recompute db caches")
			}

			if db.Close() != nil {
				fmt.Println("Cannot close DB connection:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(migrateCmd)
}

func recomputeDBCaches(gormDB *database.DB) error {
	return database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
		fmt.Print("Schedule the propagations\n")
		store.ScheduleGroupsAncestorsPropagation()
		store.SchedulePropagation([]string{"items_ancestors"})
		store.SchedulePropagation([]string{"permissions"})
		store.SchedulePropagation([]string{"results"})

		return nil
	})
}
