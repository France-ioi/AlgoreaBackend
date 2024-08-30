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

// nolint:gosec
func init() { //nolint:gochecknoinits,gocyclo
	var delay time.Duration

	deleteTempUsersCmd := &cobra.Command{
		Use:   "delete-temp-users [environment]",
		Short: "delete all temporary users with expired sessions",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if delay < 0 {
				fmt.Println("delay must be positive or equal to 0")
				os.Exit(1)
			}

			if delay > 100*365*24*time.Hour {
				fmt.Println("delay must be less than 100 years")
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
			if err != nil {
				fmt.Println("Fatal error: ", err)
				os.Exit(1)
			}

			err = database.NewDataStore(application.Database).Users().DeleteTemporaryWithTraps(delay)
			if err != nil {
				fmt.Println("Fatal error: ", err)
				os.Exit(1)
			}

			// Success
			fmt.Println("DONE")
		},
	}

	deleteTempUsersCmd.Flags().DurationVar(&delay, "delay", 60*24*time.Hour,
		"delay between expiration and deletion of temporary users")

	rootCmd.AddCommand(deleteTempUsersCmd)
}
