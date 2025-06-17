package cmd

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func init() { //nolint:gochecknoinits
	recomputeCmd := &cobra.Command{
		Use:   "db-recompute [environment]",
		Short: "recompute db caches",
		Long:  `recompute runs recalculation of db caches (groups ancestors, items ancestors, cached permissions, attempt results)`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()

			// open DB
			application, err := app.New()
			defer func() {
				if application != nil && application.Database != nil {
					_ = application.Database.Close()
				}
			}()
			if err != nil {
				return err
			}

			if err := recomputeDBCaches(cmd, application.Database); err != nil {
				return fmt.Errorf("cannot recompute db caches: %v", err)
			}

			// Success
			cmd.Println("DONE")

			return nil
		},
	}

	rootCmd.AddCommand(recomputeCmd)
}

func recomputeDBCaches(cmd *cobra.Command, gormDB *database.DB) error {
	return database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
		cmd.Print("Recalculating groups ancestors\n")
		if err := store.GroupGroups().CreateNewAncestors(); err != nil {
			return fmt.Errorf("cannot compute groups_groups: %v", err)
		}
		cmd.Print("Recalculating items ancestors\n")
		if err := store.ItemItems().CreateNewAncestors(); err != nil {
			return fmt.Errorf("cannot compute items_items: %v", err)
		}
		cmd.Print("Running propagation of permissions and results\n")
		store.SchedulePermissionsPropagation()
		store.ScheduleResultsPropagation()
		return nil
	})
}
