package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/France-ioi/mapstructure"
	"github.com/spf13/viper"
	assertlib "github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

func init() { //nolint:gochecknoinits
	appenv.SetDefaultEnvToTest()
}

var devEnv = "dev"

func TestLoadConfigFrom(t *testing.T) {
	assert := assertlib.New(t)

	// the test environment doesn't allow the merge of the config with a main config file for security reasons
	// so here we mock the function that returns the current environment, because we want to test the merge
	// of the config with the main config file
	monkey.Patch(appenv.Env, func() string { return devEnv })
	monkey.Patch(appenv.IsEnvTest, func() bool { return false })
	defer monkey.UnpatchAll()

	// create a temp config dir
	tmpDir, deferFunc := createTmpDir("conf-*", assert)
	defer deferFunc()

	// create a temp config file
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("server:\n  port: 1234\n"), 0o600)
	assert.NoError(err)

	// change default config values
	err = os.WriteFile(tmpDir+"/config.dev.yaml", []byte("server:\n  rootpath: '/test/'"), 0o600)
	assert.NoError(err)

	t.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "999")
	conf := loadConfigFrom("config", tmpDir)

	// test config override
	assert.EqualValues(1234, conf.Sub(serverConfigKey).GetInt("port"))

	// test env variables
	assert.EqualValues(999, conf.GetInt("server.WriteTimeout")) // does not work with Sub!

	// test 'test' section
	assert.EqualValues("/test/", conf.Sub(serverConfigKey).GetString("rootPath"))

	// test live env changes
	t.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "777")
	assert.EqualValues(777, conf.GetInt("server.WriteTimeout"))
}

func TestLoadConfigFrom_ShouldLogWarningWhenNonTestEnvAndNoMainConfigFile(t *testing.T) {
	assert := assertlib.New(t)

	origStdErr := os.Stderr
	stdErrReader, stdErrWriter, _ := os.Pipe()
	os.Stderr = stdErrWriter
	defer func() {
		os.Stderr = origStdErr
		_ = stdErrWriter.Close()
		_ = stdErrReader.Close()
	}()

	monkey.Patch(appenv.Env, func() string { return devEnv })
	monkey.Patch(appenv.IsEnvTest, func() bool { return false })
	defer monkey.UnpatchAll()

	// create a temp config file
	tmpFile, deferFunc := createTmpFile("config-*.dev.yaml", assert)
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-8] // strip the ".dev.yaml"

	conf := loadConfigFrom(configName, os.TempDir())
	assert.NotNil(conf)

	_ = stdErrWriter.Close()
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, stdErrReader)

	assert.Contains(buf.String(), "Cannot read the main config file, ignoring it")
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

func TestLoadConfigFrom_MustNotUseMainConfigFileInTestEnv(t *testing.T) {
	assert := assertlib.New(t)
	appenv.ForceTestEnv() // to ensure it tries to find the config.test file

	// create a temp dir to hold the config files
	tmpDir, deferFunc := createTmpDir("conf-*", assert)
	defer deferFunc()

	// create a main config file inside the tmp dir, and define two distinct yaml parameters in it
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("param1: 1\nparam2: 2"), 0o600)
	assert.NoError(err)

	// create a temp test config file inside the tmp dir, and define only one of the two parameters in it
	err = os.WriteFile(tmpDir+"/config.test.yaml", []byte("param1: 3"), 0o600)
	assert.NoError(err)

	conf := loadConfigFrom("config", tmpDir)
	assert.NotNil(conf)

	// the config of the test file should be used, and the one in the main file should not be used at all
	assert.EqualValues(3, conf.GetInt("param1"))
	assert.False(conf.IsSet("param2"))
}

func TestLoadConfigFrom_ShouldCrashIfTestEnvAndConfigTestNotPresent(t *testing.T) {
	assert := assertlib.New(t)
	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config dir
	tmpDir, deferFunc := createTmpDir("conf-*", assert)
	defer deferFunc()

	// create a temp config file
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("param1: 1"), 0o600)
	assert.NoError(err)

	assert.Panics(func() {
		_ = loadConfigFrom("config", tmpDir)
	})
}

func TestLoadConfigFrom_IgnoresEnvConfigFileIfMissing(t *testing.T) {
	assert := assertlib.New(t)

	// the test environment doesn't allow the merge of the config with a main config file for security reasons
	// so here we mock the function that returns the current environment, because we want to test the merge
	// of the config with the main config file
	monkey.Patch(appenv.Env, func() string { return devEnv })
	monkey.Patch(appenv.IsEnvTest, func() bool { return false })
	defer monkey.UnpatchAll()

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
	globalConfig.Set("database.collation", "stuff")
	t.Setenv("ALGOREA_DATABASE__TLSCONFIG", "v99") // env var which was not defined before
	dbConfig, err := DBConfig(globalConfig)
	assert.NoError(err)
	assert.Equal("stuff", dbConfig.Collation)
	assert.Equal("v99", dbConfig.TLSConfig)
}

func TestDBConfig_UnmarshallingError(t *testing.T) {
	// don't know if it is really possible to get this error
	assert := assertlib.New(t)
	globalConfig := viper.New()
	monkey.PatchInstanceMethod(reflect.TypeOf(globalConfig), "Unmarshal",
		func(_ *viper.Viper, _ interface{}, _ ...viper.DecoderConfigOption) error {
			return fmt.Errorf("unmarshalling error")
		},
	)
	defer monkey.UnpatchAll()
	_, err := DBConfig(globalConfig)
	assert.EqualError(err, "unmarshalling error")
}

func TestDBConfig_StructToMapError(t *testing.T) {
	// unexpected error, must monkey patch it
	assert := assertlib.New(t)
	globalConfig := viper.New()
	monkey.Patch(mapstructure.Decode, func(_ interface{}, _ interface{}) error {
		return fmt.Errorf("struct2map error")
	})
	defer monkey.UnpatchAll()
	_, err := DBConfig(globalConfig)
	assert.EqualError(err, "struct2map error")
}

func TestTokenConfig_Success(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	monkey.Patch(token.BuildConfig, func(_ *viper.Viper) (*token.Config, error) {
		return &token.Config{PlatformName: "test"}, nil
	})
	defer monkey.UnpatchAll()
	config, err := TokenConfig(globalConfig)
	assert.NoError(err)
	assert.Equal("test", config.PlatformName)
}

func TestTokenConfig_Error(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("token.PublicKeyFile", "notafile")
	_, err := TokenConfig(globalConfig)
	assert.Contains(err.Error(), "no such file or directory")
}

func TestAuthConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("auth.anykey", 42)
	config := AuthConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_AUTH__ANYKEY", "999")
	assert.Equal(999, config.GetInt("anykey"))
}

func TestLoggingConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("logging.anykey", 42)
	config := LoggingConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_LOGGING__ANYKEY", "999")
	assert.Equal(999, config.GetInt("anykey"))
}

func TestServerConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("server.anykey", 42)
	config := ServerConfig(globalConfig)
	assert.Equal(42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_SERVER__ANYKEY", "999")
	assert.Equal(999, config.GetInt("anykey"))
}

func TestDomainsConfig_Success(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	sampleDomain := domain.ConfigItem{
		Domains:           []string{"localhost", "other"},
		AllUsersGroup:     2,
		TempUsersGroup:    3,
		NonTempUsersGroup: 4,
	}
	globalConfig.Set("domains", []domain.ConfigItem{sampleDomain})
	config, err := DomainsConfig(globalConfig)
	assert.NoError(err)
	assert.Len(config, 1)
	assert.Equal(sampleDomain, config[0])
}

func TestDomainsConfig_Empty(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []string{})
	config, err := DomainsConfig(globalConfig)
	assert.NoError(err)
	assert.Len(config, 0)
}

func TestDomainsConfig_Error(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []int{1, 2})
	_, err := DomainsConfig(globalConfig)
	assert.EqualError(err, "2 error(s) decoding:\n\n* '[0]' expected a map, got 'int'\n* '[1]' expected a map, got 'int'")
}

func TestReplaceAuthConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("auth.ClientID", "42")
	application, err := New()
	assert.NoError(err)
	application.ReplaceAuthConfig(globalConfig)
	assert.Equal("42", application.Config.Get("auth.ClientID"))
	// not tested: that it is been pushed to the API
}

func TestReplaceDomainsConfig(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []map[string]interface{}{{"domains": []string{"localhost", "other"}}})
	application, _ := New()
	application.ReplaceDomainsConfig(globalConfig)
	expected := []domain.ConfigItem{{
		Domains:           []string{"localhost", "other"},
		AllUsersGroup:     0,
		TempUsersGroup:    0,
		NonTempUsersGroup: 0,
	}}
	config, _ := DomainsConfig(application.Config)
	assert.Equal(expected, config)
	// not tested: that it is been pushed to the API
}

func TestReplaceDomainsConfig_Panic(t *testing.T) {
	assert := assertlib.New(t)
	globalConfig := viper.New()
	globalConfig.Set("domains", []int{1, 2})
	application := &Application{Config: viper.New()}
	assert.Panics(func() {
		application.ReplaceDomainsConfig(globalConfig)
	})
}

func Test_configDirectory_StripsOnlyTheLastOccurrenceOfApp(t *testing.T) {
	monkey.Patch(os.Getwd, func() (string, error) { return "/app/something/app/ab/app/token", nil })
	defer monkey.UnpatchAll()
	dir := configDirectory()
	assertlib.Equal(t, "/app/something/app/ab/conf", dir)
}

func createTmpFile(pattern string, assert *assertlib.Assertions) (tmpFile *os.File, deferFunc func()) {
	// create a temp config file
	tmpFile, err := os.CreateTemp(os.TempDir(), pattern)
	assert.NoError(err)
	return tmpFile, func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}
}

func createTmpDir(pattern string, assert *assertlib.Assertions) (name string, deferFun func()) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), pattern)
	assert.NoError(err)
	return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
}
