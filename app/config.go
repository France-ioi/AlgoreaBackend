package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/France-ioi/mapstructure"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

const (
	defaultConfigName = "config"
	envPrefix         = "algorea"
)

// Configurations keys are sub configuration that can be fetched.
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
// 6) default.
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

	// If we are in test environment, we do not want to read the main config file,
	// because it might contain the credentials to a live database, and running the tests will erase the database
	if !appenv.IsEnvTest() {
		if err = config.ReadInConfig(); err != nil {
			fmt.Fprint(os.Stderr, "Cannot read the main config file, ignoring it: ", err)
		}
	}

	environment := appenv.Env()
	fmt.Fprintf(os.Stderr, "Loading environment: %s\n", environment)

	config.SetConfigName(filename + "." + environment)
	if err = config.MergeInConfig(); err != nil {
		if appenv.IsEnvTest() {
			fmt.Fprintf(os.Stderr, "Cannot read the %q config file: %s", environment, err)
			panic("Cannot read the test config file")
		}
		fmt.Fprintf(os.Stderr, "Cannot merge %q config file, ignoring it: %s", environment, err)
	}

	return config
}

var configPathTestRegexp = regexp.MustCompile(`.*([/\\]app(?:[/\\][a-z]+)*?)$`)

func configDirectory() string {
	cwd, _ := os.Getwd()
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		match := configPathTestRegexp.FindStringSubmatchIndex(cwd)
		if match != nil {
			cwd = cwd[:match[2]]
		}
	}
	return filepath.Dir(cwd + "/conf/")
}

// ReplaceAuthConfig replaces the auth part of the config by the given one.
// loggerOptional is an optional logger to use, if not provided, a new logger will be created from the config.
func (app *Application) ReplaceAuthConfig(newGlobalConfig *viper.Viper, loggerOptional ...*logging.Logger) {
	app.Config.Set(authConfigKey, newGlobalConfig.Get(authConfigKey))
	_ = app.Reset(app.Config, loggerOptional...) // cannot return an error in this case
}

// ReplaceDomainsConfig replaces the domains part of the config by the given one.
// loggerOptional is an optional logger to use, if not provided, a new logger will be created from the config.
func (app *Application) ReplaceDomainsConfig(newGlobalConfig *viper.Viper, loggerOptional ...*logging.Logger) {
	app.Config.Set(domainsConfigKey, newGlobalConfig.Get(domainsConfigKey))
	if err := app.Reset(app.Config, loggerOptional...); err != nil {
		panic(err)
	}
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
func DBConfig(globalConfig *viper.Viper) (config *mysql.Config, err error) {
	// Env variables are not loaded if the keys do not exist in the config file
	// To fix this issue, instead of loading config files and overriding with env vars,
	// we load all possible keys (with their default value), override with config files,
	// and then environmenent variables.
	emptyConfig := &map[string]interface{}{}
	if err = mapstructure.Decode(mysql.NewConfig(), &emptyConfig); err != nil {
		return // unexpected
	}
	vConfig := viper.New()
	_ = vConfig.MergeConfigMap(*emptyConfig) // the function always return nil
	if conf := globalConfig.GetStringMap(databaseConfigKey); conf != nil {
		_ = vConfig.MergeConfigMap(conf)
	}
	vConfig.SetEnvPrefix(fmt.Sprintf("%s_%s_", envPrefix, databaseConfigKey))
	vConfig.AutomaticEnv()
	err = vConfig.Unmarshal(&config)
	return
}

// TokenConfig returns the token fixed config from the global config.
func TokenConfig(globalConfig *viper.Viper) (*token.Config, error) {
	sub := subconfig(globalConfig, tokenConfigKey)
	return token.BuildConfig(sub)
}

// AuthConfig returns an auth dynamic config from the global config.
// (env var changes impacts values).
func AuthConfig(globalConfig *viper.Viper) *viper.Viper {
	return subconfig(globalConfig, authConfigKey)
}

// LoggingConfig returns a logging dynamic config from the global config
// (env var changes impacts values).
func LoggingConfig(globalConfig *viper.Viper) *viper.Viper {
	return subconfig(globalConfig, loggingConfigKey)
}

// ServerConfig returns a logging dynamic config from the global config
// (env var changes impacts values).
func ServerConfig(globalConfig *viper.Viper) *viper.Viper {
	return subconfig(globalConfig, serverConfigKey)
}

// DomainsConfig returns the domains fixed config from the global config
// Panic in case of marshaling error.
func DomainsConfig(globalConfig *viper.Viper) (config []domain.ConfigItem, err error) {
	globalConfig.SetDefault(domainsConfigKey, []interface{}{})
	// note that `.Sub` cannot be used to get a slice
	err = globalConfig.UnmarshalKey(domainsConfigKey, &config)
	return
}
