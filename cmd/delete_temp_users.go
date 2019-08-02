package cmd

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// nolint:gosec
func init() { // nolint:gochecknoinits,gocyclo

	var deleteTempUsersCmd = &cobra.Command{
		Use:   "delete-temp-users [environment]",
		Short: "delete all temporary users with expired sessions",
		Run: func(cmd *cobra.Command, args []string) {
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

			err = database.NewDataStore(application.Database).Users().DeleteTemporaryWithTraps()
			if err != nil {
				fmt.Println("Fatal error: ", err)
				os.Exit(1)
			}

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(deleteTempUsersCmd)
}
