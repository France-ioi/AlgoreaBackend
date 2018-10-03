package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config is the application configuration
var Config *viper.Viper

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

	Config = viper.New()

	loadDefaults()

	// through env variables
	Config.SetEnvPrefix("algorea") // env variables must be prefixed by "ALGOREA_"
	Config.AutomaticEnv()          // read in environment variables

	// through the config file
	Config.SetConfigFile(ConfigFile)
	if err := Config.ReadInConfig(); err != nil {
		fmt.Println("Cannot read config:", err)
		os.Exit(1)
	}

}

func loadDefaults() {

	// server
	Config.SetDefault("server.port", 8080)
	Config.SetDefault("server.read_timeout", 60)  // in seconds
	Config.SetDefault("server.write_timeout", 60) // in seconds
}
