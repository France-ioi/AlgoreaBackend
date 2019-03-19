package config

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// Database is the part of the config related to the database
type Database struct {
	Connection mysql.Config
}

// Server is the part of the config for the HTTP Server
type Server struct {
	Port         int32
	ReadTimeout  int32
	WriteTimeout int32
	RootPath     string
}

// ReverseProxy is the part of the config for the Reverse Proxy
type ReverseProxy struct {
	Server string
}

// Logging for all config related to logger
type Logging struct {
	Format        string
	Output        string
	Level         string
	LogSQLQueries bool
}

// Auth is the part of the config related to the user authentication
type Auth struct {
	ProxyURL string
}

// Root is the root of the app configuration
type Root struct {
	Server       Server
	Database     Database
	ReverseProxy ReverseProxy
	Timeout      int32
	Logging      Logging
	Auth         Auth
}

var (
	configName = "config"
	configDir  = configDirectory()
)

// Load loads the app configuration from files, flags, env, ..., and maps it to the config struct
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func Load() (*Root, error) {
	var err error

	var config *Root
	viperConfig := viper.New()

	setDefaults(viperConfig)

	// through env variables
	viperConfig.SetEnvPrefix("algorea") // env variables must be prefixed by "ALGOREA_"
	viperConfig.AutomaticEnv()          // read in environment variables

	// through the config file
	viperConfig.SetConfigName(configName)
	viperConfig.AddConfigPath(configDir)

	if err = viperConfig.ReadInConfig(); err != nil {
		log.Print("Cannot read the config file, ignoring it.")
	}

	// map the given config to a static struct
	if err = viperConfig.Unmarshal(&config); err != nil {
		log.Fatal("Cannot map the given config to the expected configuration struct:", err)
		return nil, err
	}
	return config, nil
}

func setDefaults(c *viper.Viper) {

	// root
	c.SetDefault("timeout", 15)

	// server
	c.SetDefault("server.port", 8080)
	c.SetDefault("server.readTimeout", 60)  // in seconds
	c.SetDefault("server.writeTimeout", 60) // in seconds
	c.SetDefault("server.rootpath", "/")

	// logging
	c.SetDefault("logging.format", "json")
	c.SetDefault("logging.output", "file")
	c.SetDefault("logging.level", "info")
	c.SetDefault("logging.logSqlQueries", true)

	// reverse proxy
	c.SetDefault("reverseproxy.server", "http://localhost:3000")

}

func configDirectory() string {
	_, codeFilePath, _, _ := runtime.Caller(0)
	codeDir := filepath.Dir(codeFilePath)
	return filepath.Dir(codeDir + "/../../conf/")
}
