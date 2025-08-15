package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/France-ioi/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
	"github.com/France-ioi/AlgoreaBackend/v2/testhelpers/testoutput"
)

func init() { //nolint:gochecknoinits
	appenv.SetDefaultEnvToTest()
}

const devEnv = "dev"

func TestLoadConfigFrom(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// the test environment doesn't allow the merge of the config with a main config file for security reasons
	// so here we mock the function that returns the current environment, because we want to test the merge
	// of the config with the main config file
	monkey.Patch(appenv.Env, func() string { return devEnv })
	monkey.Patch(appenv.IsEnvTest, func() bool { return false })
	defer monkey.UnpatchAll()

	// create a temp config dir
	tmpDir := t.TempDir()

	// create a temp config file
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("server:\n  port: 1234\n"), 0o600)
	require.NoError(t, err)

	// change default config values
	err = os.WriteFile(tmpDir+"/config.dev.yaml", []byte("server:\n  rootpath: '/test/'"), 0o600)
	require.NoError(t, err)

	t.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "999")
	conf := loadConfigFrom("config", tmpDir)
	require.NotNil(t, conf)

	// test config override
	assert.EqualValues(t, 1234, conf.Sub(serverConfigKey).GetInt("port"))

	// test env variables
	assert.EqualValues(t, 999, conf.GetInt("server.WriteTimeout")) // does not work with Sub!

	// test 'test' section
	assert.EqualValues(t, "/test/", conf.Sub(serverConfigKey).GetString("rootPath"))

	// test live env changes
	t.Setenv("ALGOREA_SERVER__WRITETIMEOUT", "777")
	assert.EqualValues(t, 777, conf.GetInt("server.WriteTimeout"))
}

func TestLoadConfigFrom_ShouldLogWarningWhenNonTestEnvAndNoMainConfigFile(t *testing.T) {
	testoutput.SuppressIfPasses(t)

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
	tmpFile, deferFunc := createTmpFile(t, "config-*.dev.yaml")
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-8] // strip the ".dev.yaml"

	conf := loadConfigFrom(configName, os.TempDir())
	require.NotNil(t, conf)

	_ = stdErrWriter.Close()
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, stdErrReader)

	assert.Contains(t, buf.String(), "Cannot read the main config file, ignoring it")
}

func TestLoadConfigFrom_IgnoresMainConfigFileIfMissing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config file
	tmpFile, deferFunc := createTmpFile(t, "config-*.test.yaml")
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-10] // strip the ".test.yaml"

	conf := loadConfigFrom(configName, os.TempDir())
	assert.NotNil(t, conf)
}

func TestLoadConfigFrom_MustNotUseMainConfigFileInTestEnv(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	appenv.ForceTestEnv() // to ensure it tries to find the config.test file

	// create a temp dir to hold the config files
	tmpDir := t.TempDir()

	// create a main config file inside the tmp dir, and define two distinct yaml parameters in it
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("param1: 1\nparam2: 2"), 0o600)
	require.NoError(t, err)

	// create a temp test config file inside the tmp dir, and define only one of the two parameters in it
	err = os.WriteFile(tmpDir+"/config.test.yaml", []byte("param1: 3"), 0o600)
	require.NoError(t, err)

	conf := loadConfigFrom("config", tmpDir)
	require.NotNil(t, conf)

	// the config of the test file should be used, and the one in the main file should not be used at all
	assert.EqualValues(t, 3, conf.GetInt("param1"))
	assert.False(t, conf.IsSet("param2"))
}

func TestLoadConfigFrom_ShouldCrashIfTestEnvAndConfigTestNotPresent(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	appenv.SetDefaultEnvToTest() // to ensure it tries to find the config.test file

	// create a temp config dir
	tmpDir := t.TempDir()

	// create a temp config file
	err := os.WriteFile(tmpDir+"/config.yaml", []byte("param1: 1"), 0o600)
	require.NoError(t, err)

	assert.Panics(t, func() {
		_ = loadConfigFrom("config", tmpDir)
	})
}

func TestLoadConfigFrom_IgnoresEnvConfigFileIfMissing(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	// the test environment doesn't allow the merge of the config with a main config file for security reasons
	// so here we mock the function that returns the current environment, because we want to test the merge
	// of the config with the main config file
	monkey.Patch(appenv.Env, func() string { return devEnv })
	monkey.Patch(appenv.IsEnvTest, func() bool { return false })
	defer monkey.UnpatchAll()

	// create a temp config file
	tmpFile, deferFunc := createTmpFile(t, "config-*.yaml")
	defer deferFunc()

	fileName := filepath.Base(tmpFile.Name())
	configName := fileName[:len(fileName)-5] // strip the ".yaml"

	conf := loadConfigFrom(configName, os.TempDir())

	assert.NotNil(t, conf)
}

func TestLoadConfig_Concurrent(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	t.Setenv("ALGOREA_ENV", "")
	_ = os.Unsetenv("ALGOREA_ENV")
	appenv.SetDefaultEnvToTest()

	assert.NotPanics(t, func() { LoadConfig() })
	const numGoroutines = 1000
	done := make(chan struct{}, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			assert.NotPanics(t, func() { LoadConfig() })
			done <- struct{}{}
		}()
	}
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestDBConfig_Success(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("database.collation", "stuff")
	t.Setenv("ALGOREA_DATABASE__TLSCONFIG", "v99") // env var which was not defined before
	dbConfig, err := DBConfig(globalConfig)
	require.NoError(t, err)
	assert.Equal(t, "stuff", dbConfig.Collation)
	assert.Equal(t, "v99", dbConfig.TLSConfig)
}

func TestDBConfig_UnmarshallingError(t *testing.T) {
	// don't know if it is really possible to get this error
	globalConfig := viper.New()
	monkey.PatchInstanceMethod(reflect.TypeOf(globalConfig), "Unmarshal",
		func(_ *viper.Viper, _ interface{}, _ ...viper.DecoderConfigOption) error {
			return errors.New("unmarshalling error")
		},
	)
	defer monkey.UnpatchAll()
	_, err := DBConfig(globalConfig)
	assert.EqualError(t, err, "unmarshalling error")
}

func TestDBConfig_StructToMapError(t *testing.T) {
	// unexpected error, must monkey patch it
	globalConfig := viper.New()
	monkey.Patch(mapstructure.Decode, func(_ interface{}, _ interface{}) error {
		return errors.New("struct2map error")
	})
	defer monkey.UnpatchAll()
	_, err := DBConfig(globalConfig)
	assert.EqualError(t, err, "struct2map error")
}

func TestTokenConfig_Success(t *testing.T) {
	globalConfig := viper.New()
	monkey.Patch(token.BuildConfig, func(_ *viper.Viper) (*token.Config, error) {
		return &token.Config{PlatformName: "test"}, nil
	})
	defer monkey.UnpatchAll()
	config, err := TokenConfig(globalConfig)
	require.NoError(t, err)
	assert.Equal(t, "test", config.PlatformName)
}

func TestTokenConfig_Error(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("token.PublicKeyFile", "notafile")
	_, err := TokenConfig(globalConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestAuthConfig(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("auth.anykey", 42)
	config := AuthConfig(globalConfig)
	require.NotNil(t, config)
	assert.Equal(t, 42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_AUTH__ANYKEY", "999")
	assert.Equal(t, 999, config.GetInt("anykey"))
}

func TestLoggingConfig(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("logging.anykey", 42)
	config := LoggingConfig(globalConfig)
	require.NotNil(t, config)
	assert.Equal(t, 42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_LOGGING__ANYKEY", "999")
	assert.Equal(t, 999, config.GetInt("anykey"))
}

func TestServerConfig(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("server.anykey", 42)
	config := ServerConfig(globalConfig)
	require.NotNil(t, config)
	assert.Equal(t, 42, config.GetInt("anykey"))
	t.Setenv("ALGOREA_SERVER__ANYKEY", "999")
	assert.Equal(t, 999, config.GetInt("anykey"))
}

func TestDomainsConfig_Success(t *testing.T) {
	globalConfig := viper.New()
	sampleDomain := domain.ConfigItem{
		Domains:           []string{"localhost", "other"},
		AllUsersGroup:     2,
		TempUsersGroup:    3,
		NonTempUsersGroup: 4,
	}
	globalConfig.Set("domains", []domain.ConfigItem{sampleDomain})
	config, err := DomainsConfig(globalConfig)
	require.NoError(t, err)
	assert.Len(t, config, 1)
	assert.Equal(t, sampleDomain, config[0])
}

func TestDomainsConfig_Empty(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("domains", []string{})
	config, err := DomainsConfig(globalConfig)
	require.NoError(t, err)
	assert.Empty(t, config)
}

func TestDomainsConfig_Error(t *testing.T) {
	globalConfig := viper.New()
	globalConfig.Set("domains", []int{1, 2})
	_, err := DomainsConfig(globalConfig)
	assert.EqualError(t, err, "2 error(s) decoding:\n\n* '[0]' expected a map, got 'int'\n* '[1]' expected a map, got 'int'")
}

func TestReplaceAuthConfig(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	mockDatabaseOpen()
	defer monkey.UnpatchAll()

	globalConfig := viper.New()
	globalConfig.Set("auth.ClientID", "42")
	logger, _ := logging.NewMockLogger()
	application, err := New(logger)
	require.NoError(t, err)
	application.ReplaceAuthConfig(globalConfig, logger)
	assert.Equal(t, "42", application.Config.Get("auth.ClientID"))
	// not tested: that it is been pushed to the API
}

func TestReplaceDomainsConfig(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	mockDatabaseOpen()
	defer monkey.UnpatchAll()

	globalConfig := viper.New()
	globalConfig.Set("domains", []map[string]interface{}{{"domains": []string{"localhost", "other"}}})
	logger, _ := logging.NewMockLogger()
	application, _ := New(logger)
	application.ReplaceDomainsConfig(globalConfig, logger)
	expected := []domain.ConfigItem{{
		Domains:           []string{"localhost", "other"},
		AllUsersGroup:     0,
		TempUsersGroup:    0,
		NonTempUsersGroup: 0,
	}}
	config, _ := DomainsConfig(application.Config)
	assert.Equal(t, expected, config)
	// not tested: that it is been pushed to the API
}

func TestReplaceDomainsConfig_Panic(t *testing.T) {
	testoutput.SuppressIfPasses(t)

	mockDatabaseOpen()
	defer monkey.UnpatchAll()

	globalConfig := viper.New()
	globalConfig.Set("domains", []int{1, 2})
	application := &Application{Config: viper.New()}
	assert.Panics(t, func() {
		application.ReplaceDomainsConfig(globalConfig)
	})
}

func Test_configDirectory_StripsOnlyTheLastOccurrenceOfApp(t *testing.T) {
	monkey.Patch(os.Getwd, func() (string, error) { return "/app/something/app/ab/app/token", nil })
	defer monkey.UnpatchAll()
	dir := configDirectory()
	assert.Equal(t, "/app/something/app/ab/conf", dir)
}

func createTmpFile(t *testing.T, pattern string) (tmpFile *os.File, deferFunc func()) {
	t.Helper()

	// create a temp config file
	tmpFile, err := os.CreateTemp(os.TempDir(), pattern)
	require.NoError(t, err)
	return tmpFile, func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}
}
