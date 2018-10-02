package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// ConfigFile defines the file name which will be used to read the configuration
var ConfigFile = "conf/default.yaml"

// LoadConfig does load the app configuration in viper
func LoadConfig() {

	viper.SetConfigFile(ConfigFile)

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Cannot read config:", err)
		os.Exit(1)
	}
}
