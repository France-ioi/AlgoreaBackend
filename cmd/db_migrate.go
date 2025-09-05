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
			var applied int
			for {
				err = goose.UpByOneContext(ctx, db, "db/migrations")
				if errors.Is(err, goose.ErrNoNextVersion) {
					break
				}
				if err != nil {
					return fmt.Errorf("unable to apply migration: %w", err)
				}
				applied++
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
