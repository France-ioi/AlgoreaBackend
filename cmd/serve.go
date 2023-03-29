// Package cmd provides the app's commands available through command line.
package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
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
			}

			log.Println("Starting application: environment =", appenv.Env())

			var application *app.Application
			application, err = app.New()
			if err != nil {
				log.Fatal(err)
			}

			if !skipChecks {
				err = application.CheckConfig()
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
