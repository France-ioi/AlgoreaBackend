package app

import (
	"fmt"
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

const (
	defaultConfigName = "config"
	envPrefix         = "algorea"
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

// LoadConfig loads and return the global configuration from files, flags, env, ...
// The precedence of config values in viper is the following:
// 1) explicit call to Set
// 2) flag (i.e., settable by command line)
// 3) env
// 4) config file
// 5) key/value store
// 6) default
func LoadConfig() *viper.Viper {
	return loadConfigFrom(defaultConfigName, configDirectory())
}

func loadConfigFrom(filename, directory string) *viper.Viper {
	var err error

	config := viper.New()

	// through env variables
	config.SetEnvPrefix(envPrefix)                           // env variables must be prefixed by "ALGOREA_"
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "__")) // substructs use "__" as key separator
	config.AutomaticEnv()                                    // read in environment variables

	// through the config file
	config.SetConfigName(filename)
	config.AddConfigPath(directory)

	if err = config.ReadInConfig(); err != nil {
		log.Print("Cannot read the main config file, ignoring it: ", err)
	}

	environment := appenv.Env()
	log.Printf("Loading environment: %s\n", environment)

	config.SetConfigName(filename + "." + environment)
	if err = config.MergeInConfig(); err != nil {
		log.Printf("Cannot merge %q config file, ignoring it: %s", environment, err)
	}

	return config
}

func configDirectory() string {
	_, codeFilePath, _, _ := runtime.Caller(0)
	codeDir := filepath.Dir(codeFilePath)
	return filepath.Dir(codeDir + "/../conf/")
}

// ReplaceAuthConfig replaces the auth part of the config by the given one.
func (app *Application) ReplaceAuthConfig(newGlobalConfig *viper.Viper) {
	app.apiCtx.SetAuthConfig(AuthConfig(newGlobalConfig))
	app.Config.Set(authConfigKey, newGlobalConfig.Get(authConfigKey))
}

// ReplaceDomainsConfig replaces the domains part of the config by the given one.
func (app *Application) ReplaceDomainsConfig(newGlobalConfig *viper.Viper) {
	app.apiCtx.DomainConfig = DomainsConfig(newGlobalConfig)
	app.Config.Set(domainsConfigKey, newGlobalConfig.Get(domainsConfigKey))
}

//
// Subconfig accessors
// These are public and working on the viper global config (as arg) to allow usage in app.New and every tests
//

// submapconfig returns the subconfig from a global config
// only for configuration which are maps
// Note: (kind of bug) for unmarshaled configs such as db, the config item needs to be set first (e.g., by
// config file) for the env values to be read by viper

func subconfig(globalConfig *viper.Viper, subconfigKey string) *viper.Viper {
	globalConfig.SetDefault(subconfigKey, map[string]interface{}{})
	subConfig := globalConfig.Sub(subconfigKey)
	subConfig.SetEnvPrefix(fmt.Sprintf("%s_%s_", envPrefix, subconfigKey))
	subConfig.AutomaticEnv()
	return subConfig
}

// DBConfig returns the db connection fixed config from the global config.
// Panic in case of unmarshaling error
func DBConfig(globalConfig *viper.Viper) (config *mysql.Config) {
	sub := subconfig(globalConfig, databaseConfigKey)
	if err := sub.Unmarshal(&config); err != nil {
		panic("Unable to load the 'database' configuration")
	}
	return
}

// TokenConfig returns the token fixed config from the global config
// Panic in case of unmarshallign error
func TokenConfig(globalConfig *viper.Viper) *token.Config {
	sub := subconfig(globalConfig, tokenConfigKey)
	config, err := token.Initialize(sub)
	if err != nil {
		panic("Unable to load the 'token' configuration")
	}
	return config
}

// AuthConfig returns an auth dynamic config from the global config
// (env var changes impacts values)
func AuthConfig(globalConfig *viper.Viper) *viper.Viper {
	return subconfig(globalConfig, authConfigKey)
}

// LoggingConfig returns a logging dynamic config from the global config
// (env var changes impacts values)
func LoggingConfig(globalConfig *viper.Viper) *viper.Viper {
	return subconfig(globalConfig, loggingConfigKey)
}

// ServerConfig returns a logging dynamic config from the global config
// (env var changes impacts values)
func ServerConfig(globalConfig *viper.Viper) *viper.Viper {
	sub := subconfig(globalConfig, serverConfigKey)
	sub.SetDefault("rootpath", "/")
	sub.SetDefault("port", 8080)
	sub.SetDefault("readTimeout", 60)
	sub.SetDefault("writeTimeout", 60)
	return sub
}

// DomainsConfig returns the domains fixed config from the global config
// Panic in case of marshaling error.
func DomainsConfig(globalConfig *viper.Viper) (config []domain.ConfigItem) {
	globalConfig.SetDefault(domainsConfigKey, []interface{}{})
	// note that `.Sub` cannot be used to get a slice
	if err := globalConfig.UnmarshalKey(domainsConfigKey, &config); err != nil {
		panic("Unable to load 'Domains' config")
	}
	return
}
