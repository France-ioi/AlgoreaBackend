package cmd

import (
	"fmt"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// Set the environment.
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			var application *app.Application
			application, err = app.New()
			defer func() {
				if application != nil && application.Database != nil {
					_ = application.Database.Close()
				}
			}()
			if err != nil {
				return err
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
				return fmt.Errorf("error while doing propagation: %w", err)
			}

			cmd.Println("Propagation done.")

			return nil
		},
	}

	rootCmd.AddCommand(propagationCmd)
}
