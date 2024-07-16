// Package cmd provides the app's commands available through command line.
package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database/configdb"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
)

func init() { //nolint:gochecknoinits
	var skipChecks bool

	serveCmd := &cobra.Command{
		Use:   "serve [environment]",
		Short: "start http server",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// if arg given, replace the env
			if len(args) > 0 {
				appenv.SetEnv(args[0])

				if appenv.IsEnvTest() {
					log.Fatal("serve cannot be run in test environment.")
				}
			}

			log.Println("Starting application: environment =", appenv.Env())

			var application *app.Application
			application, err = app.New()
			if err != nil {
				log.Fatal(err)
			}

			if !skipChecks {
				var domainConfig []domain.ConfigItem
				domainConfig, err = app.DomainsConfig(application.Config)
				if err != nil {
					log.Fatalf("Cannot load domain config: %s\n", err)
				}
				err = configdb.CheckConfig(database.NewDataStore(application.Database), domainConfig)
				if err != nil {
					log.Fatalf("Integrity check failed: %s\nUse --skip-checks to bypass the integrity check\n", err)
				}
			}

			var server *app.Server
			server, err = app.NewServer(application)
			if err != nil {
				log.Fatal(err)
			}
			<-server.Start()
		},
	}

	serveCmd.Flags().BoolVar(&skipChecks, "skip-checks", false, "skip the integrity check at startup")
	rootCmd.AddCommand(serveCmd)
}
