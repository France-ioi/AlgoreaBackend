package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// ConfigFile defines the file name which will be used to read the configuration
var ConfigFile = "conf/default.yaml"

// LoadConfig does load the app configuration in viper
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func LoadConfig() {

	loadDefaults()

	// through env variables
	viper.SetEnvPrefix("algorea") // env variables must be prefixed by "ALGOREA_"
	viper.AutomaticEnv()          // read in environment variables

	// through the config file
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Cannot read config:", err)
		os.Exit(1)
	}

}

func loadDefaults() {

	// server
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 60)  // in seconds
	viper.SetDefault("server.write_timeout", 60) // in seconds
}
