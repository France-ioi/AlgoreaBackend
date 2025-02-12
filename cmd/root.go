package cmd

import (
	"fmt"
	"os"

	"github.com/akrylysov/algnhsa"
	_ "github.com/aws/aws-lambda-go/events" // force algnhsa dependency
	_ "github.com/aws/aws-lambda-go/lambda" // force algnhsa dependency
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
)

var rootCmd = &cobra.Command{
	Use: "AlgoreaBackend",
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.New()
		defer func() {
			if application != nil && application.Database != nil {
				_ = application.Database.Close()
			}
		}()
		if err != nil {
			return err
		}

		algnhsa.ListenAndServe(application.HTTPHandler, nil)

		return nil
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
