package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

func TestLoadConfigFrom(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "config-*.yaml")
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}()

	text := []byte("server:\n  port: 1234\n")
	_, err = tmpFile.Write(text)
	assert.NoError(err)

	// change default config values
	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-5] // strip the ".yaml"

	tmpTestFileName := tmpDir + "/" + configName + ".test.yaml"
	err = ioutil.WriteFile(tmpTestFileName, []byte("server:\n  rootpath: '/test/'"), 0644)
	assert.NoError(err)
	defer func() {
		_ = os.Remove(tmpTestFileName)
	}()

	_ = os.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "999")
	defer func() { _ = os.Unsetenv("ALGOREA_SERVER__WRITETIMEOUT") }()
	conf := loadConfigFrom(configName, tmpDir)

	// test config override
	assert.EqualValues(1234, conf.Sub(serverConfigKey).GetInt("port"))

	// test env variables
	assert.EqualValues(999, conf.GetInt("server.WriteTimeout")) // does not work with Sub!

	// test 'test' section
	assert.EqualValues("/test/", conf.Sub(serverConfigKey).GetString("RootPath"))

	// test live env changes
	_ = os.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "777")
	defer func() { _ = os.Unsetenv("ALGOREA_SERVER__WRITETIMEOUT") }()
	assert.EqualValues(777, conf.GetInt("server.WriteTimeout"))
}

func TestLoadConfigFrom_IgnoresMainConfigFileIfMissing(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpFile, deferFunc := createTmpFile("config-*.test.yaml", assert)
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-10] // strip the ".test.yaml"

	conf := loadConfigFrom(configName, os.TempDir())
	assert.NotNil(conf)
}

func TestLoadConfigFrom_IgnoresEnvConfigFileIfMissing(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpFile, deferFunc := createTmpFile("config-*.yaml", assert)
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-5] // strip the ".yaml"

	conf := loadConfigFrom(configName, os.TempDir())

	assert.NotNil(conf)
}

func TestLoadConfig_Concurrent(t *testing.T) {
	_ = os.Unsetenv("ALGOREA_ENV")
	appenv.SetDefaultEnvToTest()
	assert := assertlib.New(t)
	assert.NotPanics(func() {
		LoadConfig()
		for i := 0; i < 1000; i++ {
			go func() { LoadConfig() }()
		}
	})
}

func TestDBConfig_Success(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	_ = os.Setenv("", "myself")
	globalConfig.Set("database.connection.collation", "stuff")
	globalConfig.Set("database.connection.TLSConfig", "v88")
	// Still buggy, for unmarshaled config, the config needs to be set first (by config file
	// or manually) to allow setting it through env
	_ = os.Setenv("ALGOREA_DATABASE__CONNECTION__TLSCONFIG", "v99")
	defer func() { _ = os.Unsetenv("ALGOREA_DATABASE__CONNECTION__TLSCONFIG") }()
	dbConfig := DBConfig(globalConfig)
	assert.Equal("stuff", dbConfig.Collation)
	assert.Equal("v99", dbConfig.TLSConfig)
}

func TestDBConfig_Panic(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("database.connection.Timeout", "invalid")
	assert.Panics(func() {
		_ = DBConfig(globalConfig)
	})
}

func TestTokenConfig_Success(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	monkey.Patch(token.Initialize, func(config *viper.Viper) (*token.Config, error) {
		return &token.Config{PlatformName: "test"}, nil
	})
	defer monkey.UnpatchAll()
	config := TokenConfig(globalConfig)
	assert.Equal("test", config.PlatformName)
}

func TestTokenConfig_Panic(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("token.PublicKeyFile", "notafile")
	assert.Panics(func() {
		_ = TokenConfig(globalConfig)
	})
}

func TestAuthConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("auth.anykey", 42)
	config := AuthConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	_ = os.Setenv("ALGOREA_AUTH__ANYKEY", "999")
	defer func() { _ = os.Unsetenv("ALGOREA_AUTH__ANYKEY") }()
	assert.Equal(999, config.GetInt("anykey"))
}

func TestLoggingConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("logging.anykey", 42)
	config := LoggingConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	_ = os.Setenv("ALGOREA_LOGGING__ANYKEY", "999")
	defer func() { _ = os.Unsetenv("ALGOREA_LOGGING__ANYKEY") }()
	assert.Equal(999, config.GetInt("anykey"))
}

func TestServerConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("server.anykey", 42)
	config := ServerConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	_ = os.Setenv("ALGOREA_SERVER__ANYKEY", "999")
	defer func() { _ = os.Unsetenv("ALGOREA_SERVER__ANYKEY") }()
	assert.Equal(999, config.GetInt("anykey"))
	// test default values:
	assert.Equal("/", config.GetString("rootpath"))
	assert.Equal(8080, config.GetInt("port"))
	assert.Equal(60, config.GetInt("readTimeout"))
	assert.Equal(60, config.GetInt("writeTimeout"))
}

func TestDomainsConfig_Success(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	sampleDomain := domain.AppConfigItem{
		Domains:       []string{"localhost", "other"},
		RootGroup:     1,
		RootSelfGroup: 2,
		RootTempGroup: 3,
	}
	globalConfig.Set("domains", []domain.AppConfigItem{sampleDomain})
	config := DomainsConfig(globalConfig)
	assert.Len(config, 1)
	assert.Equal(sampleDomain, config[0])
}

func TestDDomainsConfig_Empty(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []string{})
	config := DomainsConfig(globalConfig)
	assert.Len(config, 0)
}

func TestDomainsConfig_Panic(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []int{1, 2})
	assert.Panics(func() {
		_ = DomainsConfig(globalConfig)
	})
}

func createTmpFile(pattern string, assert *assertlib.Assertions) (tmpFile *os.File, deferFunc func()) {
	// create a temp config file
	tmpFile, err := ioutil.TempFile(os.TempDir(), pattern)
	assert.NoError(err)
	return tmpFile, func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}
}
