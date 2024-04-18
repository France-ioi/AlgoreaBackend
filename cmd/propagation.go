package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

const (
	propagationLockName    = "listener_propagate"
	propagationLockTimeout = 600 * time.Second
)

func init() { //nolint:gochecknoinits
	propagationCmd := &cobra.Command{
		Use:   "propagation [environment]",
		Short: "apply propagation to the database",
		Long:  `runs items, permissions and results propagation`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Set the environment.
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appConfig := app.LoadConfig()

			// Connect to database.
			var db *sql.DB
			databaseConfig, err := app.DBConfig(appConfig)
			if err != nil {
				fmt.Println("Unable to load the database config: ", err)
				os.Exit(1)
			}

			db, err = database.OpenRawDBConnection(databaseConfig.FormatDSN())
			if err != nil {
				fmt.Println("Cannot open database connection: ", err)
				os.Exit(1)
			}

			// Initialize the logger to log the SQL queries.
			loggingConfig := app.LoggingConfig(appConfig)
			logging.SharedLogger.Configure(loggingConfig)

			gormDB, err := database.Open(db)
			if err != nil {
				fmt.Println("Cannot retrieve GORM instance: ", err)
				os.Exit(1)
			}

			// Propagation.
			// We use a lock because we don't want this process to be called concurrently.
			err = database.NewDataStore(gormDB).WithNamedLock(propagationLockName, propagationLockTimeout, func(s *database.DataStore) error {
				return s.InTransaction(func(store *database.DataStore) error {
					store.ScheduleItemsAncestorsPropagation()
					store.ScheduleResultsPropagation()

					return nil
				})
			})
			if err != nil {
				fmt.Println("Error while doing propagation: ", err)
				os.Exit(1)
			}

			fmt.Println("Propagation done.")

			// Close database connection.
			if db.Close() != nil {
				fmt.Println("Cannot close DB connection:", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(propagationCmd)
}
