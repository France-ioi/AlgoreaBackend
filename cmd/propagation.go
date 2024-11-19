package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

const (
	propagationCommandLockName    = "propagation_command"
	propagationCommandLockTimeout = 600 * time.Second
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

			var application *app.Application
			application, err = app.New()
			if err != nil {
				log.Fatal(err)
			}

			// Propagation.
			// We use a lock because we don't want this process to be called concurrently.
			err = database.NewDataStore(application.Database).
				WithNamedLock(propagationCommandLockName, propagationCommandLockTimeout, func(s *database.DataStore) error {
					return s.InTransaction(func(store *database.DataStore) error {
						store.SchedulePermissionsPropagation()
						store.ScheduleResultsPropagation()

						return nil
					})
				})
			if err != nil {
				fmt.Println("Error while doing propagation: ", err)
				os.Exit(1)
			}

			fmt.Println("Propagation done.")
		},
	}

	rootCmd.AddCommand(propagationCmd)
}
