package app

import (
	"log"

	"github.com/spf13/viper"
)

type serverConfig struct {
	Port         int32
	ReadTimeout  int32
	WriteTimeout int32
}

type rootConfig struct {
	Server serverConfig
}

// Config is the application configuration
var Config rootConfig

// ConfigFile defines the file name which will be used to read the configuration
var ConfigFile = "conf/default.yaml"

// LoadConfig loads the app configuration from files, flags, env, ..., and maps it to the config struct
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func (config *rootConfig) Load() error {

	setDefaults()

	// through env variables
	viper.SetEnvPrefix("algorea") // env variables must be prefixed by "ALGOREA_"
	viper.AutomaticEnv()          // read in environment variables

	// through the config file
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Cannot read config:", err)
		return err
	}

	// map the given config to a static struct
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Cannot map the given config to the expected configuration struct:", err)
		return err
	}
	return nil
}

func setDefaults() {

	// server
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", 60)  // in seconds
	viper.SetDefault("server.writeTimeout", 60) // in seconds
}
