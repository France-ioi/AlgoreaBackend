package cmd

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database/configdb"
)

func init() { //nolint:gochecknoinits
	installCmd := &cobra.Command{
		Use:   "install [environment]",
		Short: "fill the database with required data",
		Long: `If the root group IDs specified in the config file
do not exist or have missing relations, creates them all
(groups, groups_groups, and groups_ancestors)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnv("dev")

			application, err := app.New()
			defer func() {
				if application != nil && application.Database != nil {
					_ = application.Database.Close()
				}
			}()
			if err != nil {
				return err
			}

			domainsConfig, err := app.DomainsConfig(application.Config)
			if err != nil {
				return err
			}

			err = configdb.CreateMissingData(database.NewDataStore(application.Database), domainsConfig)
			if err != nil {
				return err
			}

			// Success
			fmt.Println("DONE")

			return nil
		},
	}

	rootCmd.AddCommand(installCmd)
}
