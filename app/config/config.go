package config

import (
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// Server is the part of the config for the HTTP Server
type Server struct {
	Port         int32
	ReadTimeout  int32
	WriteTimeout int32
}

// ReverseProxy is the part of the config for the Reverse Proxy
type ReverseProxy struct {
	Server string
}

// Root is the root of the app configuration
type Root struct {
	Server       Server
	Database     mysql.Config
	ReverseProxy ReverseProxy
	Timeout      int32
}

// Path defines the file name which will be used to read the configuration
var Path = "conf/default.yaml"

// Load loads the app configuration from files, flags, env, ..., and maps it to the config struct
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func Load() (*Root, error) {

	var config *Root
	setDefaults()

	// through env variables
	viper.SetEnvPrefix("algorea") // env variables must be prefixed by "ALGOREA_"
	viper.AutomaticEnv()          // read in environment variables

	// through the config file
	viper.SetConfigFile(Path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Cannot read config:", err)
		return nil, err
	}

	// map the given config to a static struct
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Cannot map the given config to the expected configuration struct:", err)
		return nil, err
	}
	return config, nil
}

func setDefaults() {

	// root
	viper.SetDefault("timeout", 15)

	// server
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", 60)  // in seconds
	viper.SetDefault("server.writeTimeout", 60) // in seconds
}
