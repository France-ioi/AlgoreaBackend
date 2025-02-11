package cmd

import (
	"fmt"
	"os"

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
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])
			}

			appenv.SetDefaultEnvToTest()

			// open DB
			application, err := app.New()
			if err != nil {
				fmt.Println("Fatal error: ", err)
				os.Exit(1)
			}

			defer func() { _ = application.Database.Close() }()

			assertNoError(recomputeDBCaches(application.Database), "Cannot recompute db caches")

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(recomputeCmd)
}

func recomputeDBCaches(gormDB *database.DB) error {
	return database.NewDataStore(gormDB).InTransaction(func(store *database.DataStore) error {
		fmt.Print("Recalculating groups ancestors\n")
		assertNoError(store.GroupGroups().CreateNewAncestors(), "Cannot compute groups_groups")
		fmt.Print("Recalculating items ancestors\n")
		assertNoError(store.ItemItems().CreateNewAncestors(), "Cannot compute items_items")
		fmt.Print("Schedule the propagations\n")
		store.SchedulePermissionsPropagation()
		store.ScheduleResultsPropagation()
		return nil
	})
}
