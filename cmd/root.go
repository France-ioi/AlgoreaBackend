package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/akrylysov/algnhsa"
	_ "github.com/aws/aws-lambda-go/events" // force algnhsa dependency
	_ "github.com/aws/aws-lambda-go/lambda" // force algnhsa dependency
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app"
)

var rootCmd = &cobra.Command{
	Use: "AlgoreaBackend",
	Run: func(cmd *cobra.Command, args []string) {
		application, err := app.New()
		if err != nil {
			log.Fatal(err)
		}

		algnhsa.ListenAndServe(application.HTTPHandler, nil)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() { //nolint:gochecknoinits
	// persistent flags will be available for every sub-commands
	// here you can bind command line flags to variables
}
