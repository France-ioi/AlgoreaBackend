package app

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

const defaultConfigName = "config"

var (
	configName = defaultConfigName
	configDir  = configDirectory()
)

// Configurations keys are sub configuration that can be fetched
const (
	serverConfigKey   string = "server"
	databaseConfigKey string = "database"
	loggingConfigKey  string = "logging"
	authConfigKey     string = "auth"
	tokenConfigKey    string = "token"
	domainsConfigKey  string = "domains"
)

// LoadConfig loads and return the app configuration from files, flags, env, ...
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func LoadConfig() *viper.Viper {
	var err error

	config := viper.New()

	// through env variables
	config.SetEnvPrefix("algorea")                           // env variables must be prefixed by "ALGOREA_"
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "__")) // substructs use "__" as key separator
	config.AutomaticEnv()                                    // read in environment variables

	// through the config file
	config.SetConfigName(configName)
	config.AddConfigPath(configDir)

	if err = config.ReadInConfig(); err != nil {
		log.Print("Cannot read the main config file, ignoring it: ", err)
	}

	environment := appenv.Env()
	log.Printf("Loading environment: %s\n", environment)

	config.SetConfigName(configName + "." + environment)
	if err = config.MergeInConfig(); err != nil {
		log.Printf("Cannot merge %q config file, ignoring it: %s", environment, err)
	}

	// Set empty map as default value to expected keys so that the subconfiguration is empty and not null
	config.SetDefault(serverConfigKey, map[string]interface{}{})
	config.SetDefault(databaseConfigKey, map[string]interface{}{})
	config.SetDefault(loggingConfigKey, map[string]interface{}{})
	config.SetDefault(authConfigKey, map[string]interface{}{})
	config.SetDefault(tokenConfigKey, map[string]interface{}{})
	config.SetDefault(domainsConfigKey, []interface{}{})

	return config
}

func configDirectory() string {
	_, codeFilePath, _, _ := runtime.Caller(0)
	codeDir := filepath.Dir(codeFilePath)
	return filepath.Dir(codeDir + "/../conf/")
}

// DBConfig returns the db connection config from the app config.
// Panic in case of error. Only for tests and commands
func DBConfig(appConfig *viper.Viper) *mysql.Config {
	var dbConfig *mysql.Config
	if err := appConfig.Sub(databaseConfigKey).Sub("connection").Unmarshal(&dbConfig); err != nil {
		panic("Unable to load the 'database' configuration")
	}
	return dbConfig
}

// TokenConfig returns the token config from the app config
// Panic in case of error. Only for tests and commands
func TokenConfig(appConfig *viper.Viper) *token.Config {
	config, err := token.Initialize(appConfig.Sub(tokenConfigKey))
	if err != nil {
		panic("Unable to load the 'token' configuration")
	}
	return config
}

// AuthConfig returns the auth config from the app config
// Only for tests and commands
func AuthConfig(appConfig *viper.Viper) *viper.Viper {
	return appConfig.Sub(authConfigKey)
}

// LoggingConfig returns the logging config from the app config
// Only for tests and commands
func LoggingConfig(appConfig *viper.Viper) *viper.Viper {
	return appConfig.Sub(loggingConfigKey)
}

// DomainsConfig returns the domains config from the app config
// Panic in case of error.
func DomainsConfig(appConfig *viper.Viper) (config []domain.AppConfigItem) {
	if err := appConfig.UnmarshalKey(domainsConfigKey, &config); err != nil {
		panic("Unable to load 'Domains' config")
	}
	return
}
