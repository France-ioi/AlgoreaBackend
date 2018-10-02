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
			server, err := app.NewServer()
			if err != nil {
				log.Fatal(err)
			}
			server.Start()
		},
	}

	rootCmd.AddCommand(serveCmd)
}
