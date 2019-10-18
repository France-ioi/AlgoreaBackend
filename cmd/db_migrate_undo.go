package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

func init() { // nolint:gochecknoinits

	var migrateUndoCmd = &cobra.Command{
		Use:   "db-migrate-undo [environment]",
		Short: "undo the last schema-change migration applied to the database",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()

			// load config
			conf := config.Load()

			// open DB
			migrations := &migrate.FileMigrationSource{Dir: "db/migrations"}
			var db *sql.DB
			databaseConfig := conf.Database.Connection
			databaseConfig.ParseTime = true
			db, err = sql.Open("mysql", databaseConfig.FormatDSN())
			if err != nil {
				fmt.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			// migrate
			var n int
			n, err = migrate.ExecMax(db, "mysql", migrations, migrate.Down, 1)
			switch {
			case err != nil:
				fmt.Println("Unable to undo a migration:", err)
				os.Exit(1)
			case n == 0:
				fmt.Println("No migrations to undo!")
			default:
				var gormDB *database.DB
				gormDB, err = database.Open(db)
				assertNoError(err, "Cannot open GORM db connection: ")
				err = database.NewDataStore(gormDB).GroupAttempts().ComputeAllGroupAttempts()
				assertNoError(err, "Cannot compute groups_attempts")
				fmt.Println("1 migration undone successfully!")
			}

			if db.Close() != nil {
				fmt.Println("Cannot close DB connection:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(migrateUndoCmd)
}
