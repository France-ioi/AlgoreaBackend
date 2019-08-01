package cmd

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
)

// nolint:gosec
func init() { // nolint:gochecknoinits,gocyclo

	var installCmd = &cobra.Command{
		Use:   "install [environment]",
		Short: "fill the database with required data",
		Long: `If the root group IDs specified in the config file
do not exist or have missing relations, creates them all
(groups, groups_groups, and groups_ancestors)`,
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
				log.Fatal(err)
			}

			err = application.CreateMissingData()
			if err != nil {
				log.Fatal(err)
			}

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(installCmd)
}
