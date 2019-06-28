package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
)

func init() { // nolint:gochecknoinits

	var serveCmd = &cobra.Command{
		Use:   "serve [environment]",
		Short: "start http server",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			var application *app.Application
			environment := "dev"
			if len(args) == 1 {
				environment = args[0]
			}

			application, err = app.New(environment)
			if err != nil {
				log.Fatal(err)
			}

			var server *app.Server
			server, err = app.NewServer(application)
			if err != nil {
				log.Fatal(err)
			}
			server.Start()
		},
	}

	rootCmd.AddCommand(serveCmd)
}
