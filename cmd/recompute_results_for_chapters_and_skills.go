package cmd

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func init() { //nolint:gochecknoinits
	recomputeResultsCmd := &cobra.Command{
		Use:   "recompute-results [environment]",
		Short: "recompute results for chapters and skills",
		Long:  `for each chapter/skill marks all results linked to it as to_be_recomputed and runs the results propagation`,
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

			store := database.NewDataStore(application.Database)
			itemNumber := 0
			err = store.Items().Where("type = 'Chapter' OR type = 'Skill'").Select("id").
				ScanAndHandleMaps(func(item map[string]interface{}) error {
					itemNumber++
					return store.InTransaction(func(store *database.DataStore) error {
						log.Printf("Recomputing results for item %s (#%d)\n", item["id"], itemNumber)
						err = store.Exec("INSERT IGNORE INTO results_recompute_for_items (item_id) values (?)", item["id"]).Error()
						if err != nil {
							return err
						}
						store.ScheduleResultsPropagation()
						return nil
					})
				}).Error()
			if err != nil {
				return fmt.Errorf("error while recomputing results: %v", err)
			}

			fmt.Println("Done.")

			return nil
		},
	}

	rootCmd.AddCommand(recomputeResultsCmd)
}
