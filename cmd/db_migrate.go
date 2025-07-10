package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
)

func init() { //nolint:gochecknoinits
	migrateCmd := &cobra.Command{
		Use:   "db-migrate [environment]",
		Short: "apply schema-change migrations to the database",
		Long:  `migrate uses go-pg migration tool under the hood supporting the same commands and an additional reset command`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				cmd.Println("Unable to load the database config: ", err)
				os.Exit(1)
			}
			databaseConfig.ParseTime = true
			db, err = sql.Open("mysql", databaseConfig.FormatDSN())
			if err != nil {
				cmd.Println("Unable to connect to the database: ", err)
				os.Exit(1)
			}

			defer func() { _ = db.Close() }()

			// migrate
			var applied int
			for {
				var appliedDuringIteration int
				appliedDuringIteration, err = migrate.ExecMax(db, "mysql", migrations, migrate.Up, 1)
				if err != nil {
					return fmt.Errorf("unable to apply migration: %w", err)
				}
				applied += appliedDuringIteration
				if appliedDuringIteration == 0 {
					break
				}
				cmd.Print(".")
			}
			cmd.Print("\n")
			switch {
			case applied == 0:
				cmd.Println("No migrations to apply!")
			default:
				cmd.Printf("%d migration(s) applied successfully!\n", applied)
			}

			return nil
		},
	}

	rootCmd.AddCommand(migrateCmd)
}
