package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	_ "github.com/France-ioi/AlgoreaBackend/v2/db/migrations" // register migrations
)

func init() { //nolint:gochecknoinits
	migrateUndoCmd := &cobra.Command{
		Use:   "db-migrate-undo [environment]",
		Short: "undo the last schema-change migration applied to the database",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()

			// open DB
			var db *sql.DB
			config := app.LoadConfig()
			databaseConfig, err := app.DBConfig(config)
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

			ctx := createContextWithLogger(config)
			err = goose.SetDialect("mysql")
			if err != nil {
				panic(err)
			}

			// migrate
			err = goose.DownContext(ctx, db, "db/migrations")
			switch {
			case errors.Is(err, goose.ErrNoMigrationFiles):
				fallthrough
			case err != nil && err.Error() == "migration 0: no current version found":
				cmd.Println("No migration to undo!")
			case err != nil:
				return fmt.Errorf("unable to undo a migration: %w", err)
			default:
				cmd.Println("1 migration undone successfully!")
			}

			return nil
		},
	}

	rootCmd.AddCommand(migrateUndoCmd)
}
