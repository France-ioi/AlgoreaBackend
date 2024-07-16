package cmd

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
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

			assertNoError(recomputeDBCaches(application.Database), "Cannot recompute db caches")

			// Success
			fmt.Println("DONE")
		},
	}

	rootCmd.AddCommand(recomputeCmd)
}
