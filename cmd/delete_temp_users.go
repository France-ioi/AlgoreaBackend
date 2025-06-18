package cmd

import (
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func init() { //nolint:gochecknoinits
	var delay time.Duration

	deleteTempUsersCmd := &cobra.Command{
		Use:   "delete-temp-users [environment]",
		Short: "delete all temporary users with expired sessions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if delay < 0 {
				cmd.Println("delay must be positive or equal to 0")
				os.Exit(1)
			}

			if delay > 100*365*24*time.Hour {
				cmd.Println("delay must be less than 100 years")
				os.Exit(1)
			}

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnv("dev")

			var application *app.Application
			var err error
			application, err = app.New()
			defer func() {
				if application != nil && application.Database != nil {
					_ = application.Database.Close()
				}
			}()
			if err != nil {
				return err
			}

			err = database.NewDataStore(application.Database).Users().DeleteTemporaryWithTraps(delay)
			if err != nil {
				return fmt.Errorf("cannot delete temporary users: %w", err)
			}

			// Success
			cmd.Println("DONE")

			return nil
		},
	}

	deleteTempUsersCmd.Flags().DurationVar(&delay, "delay", 60*24*time.Hour,
		"delay between expiration and deletion of temporary users")

	rootCmd.AddCommand(deleteTempUsersCmd)
}
