package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"gopkg.in/gorp.v1"

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
				n, err = ExecMax(db, "mysql", migrations, migrate.Up, 1)
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
		fmt.Print("Running GroupGroupStore.After()\n")
		assertNoError(store.GroupGroups().After(), "Cannot compute groups_groups") // calls createNewAncestors()
		fmt.Print("Running ItemItemStore.After()\n")
		assertNoError(store.ItemItems().After(), "Cannot compute items_items") // calls createNewAncestors() & computeAllAccess()
		return nil
	})
}

// ExecMax has been copied from the package sql-migrate.
// We need to change a variable inside the *.sql by a value in the config: the call to AWS Lambda for propagation.
//
// It would be better not to copy it like this, but do we have other options?
//
//nolint:gocognit,gocyclo
func ExecMax(db *sql.DB, dialect string, m migrate.MigrationSource, dir migrate.MigrationDirection, max int) (int, error) {
	awsAuroraPropagationTrigger := app.LoadConfig().GetString("database.aws_aurora_propagation_trigger")

	migrations, dbMap, err := migrate.PlanMigration(db, dialect, m, dir, max)
	if err != nil {
		return 0, err
	}

	// Apply migrations
	applied := 0
	for _, migration := range migrations {
		var executor migrate.SqlExecutor

		if migration.DisableTransaction {
			executor = dbMap
		} else {
			executor, err = dbMap.Begin()
			if err != nil {
				return applied, &migrate.TxError{
					Migration: migration.Migration,
					Err:       err,
				}
			}
		}

		for _, stmt := range migration.Queries {
			// Here we replace our custom variable.
			stmt = strings.Replace(stmt, "-- %%aws_aurora_propagation_trigger%%", awsAuroraPropagationTrigger, -1)

			//nolint:govet
			if _, err := executor.Exec(stmt); err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, &migrate.TxError{
					Migration: migration.Migration,
					Err:       err,
				}
			}
		}

		switch dir {
		case migrate.Up:
			err = executor.Insert(&migrate.MigrationRecord{
				Id:        migration.Id,
				AppliedAt: time.Now(),
			})
			if err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, &migrate.TxError{
					Migration: migration.Migration,
					Err:       err,
				}
			}
		case migrate.Down:
			_, err := executor.Delete(&migrate.MigrationRecord{
				Id: migration.Id,
			})
			if err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, &migrate.TxError{
					Migration: migration.Migration,
					Err:       err,
				}
			}
		default:
			panic("Not possible")
		}

		if trans, ok := executor.(*gorp.Transaction); ok {
			if err := trans.Commit(); err != nil {
				return applied, &migrate.TxError{
					Migration: migration.Migration,
					Err:       err,
				}
			}
		}

		applied++
	}

	return applied, nil
}
