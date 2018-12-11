package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

var rootCmd = &cobra.Command{
	Use: "AlgoreaBackend",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// persistent flags will be available for every sub-commands
	rootCmd.PersistentFlags().StringVar(&config.Path, "config", config.Path, "path to config file")
}
