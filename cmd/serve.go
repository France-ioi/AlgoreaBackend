package cmd

import (
  "log"

  "github.com/France-ioi/AlgoreaBackend/app"
  "github.com/spf13/cobra"
)

func init() {

  var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "start http server",
    Run: func(cmd *cobra.Command, args []string) {
      var err error

      var application *app.Application
      application, err = app.New()
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
