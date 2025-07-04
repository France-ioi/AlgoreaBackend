package cmd

import (
	"context"
	"os"

	"github.com/akrylysov/algnhsa"
	_ "github.com/aws/aws-lambda-go/events" // force algnhsa dependency
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	log "github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

var rootCmd = &cobra.Command{
	Use: "AlgoreaBackend",
	RunE: func(_ *cobra.Command, _ []string) error {
		application, err := app.New()
		closeDB := func() {
			if application != nil && application.Database != nil {
				_ = application.Database.Close()
			}
		}
		defer closeDB()
		if err != nil {
			return err
		}

		lambdaHandler := algnhsa.New(application.HTTPHandler, nil)
		lambda.StartWithOptions(lambdaHandler, lambda.WithEnableSIGTERM(func() {
			log.SharedLogger.WithContext(context.Background()).Info("Got SIGTERM, closing the DB connection")
			closeDB()
			log.SharedLogger.WithContext(context.Background()).Info("Closed the DB connection after receiving SIGTERM")
		}))

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
		os.Exit(1)
	}
}

func init() { //nolint:gochecknoinits
	// persistent flags will be available for every sub-commands
	// here you can bind command line flags to variables
}
