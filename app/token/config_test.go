package token

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/SermoDigital/jose/crypto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/tokentest"
)

func TestBuildConfig_LoadsKeysFromFile(t *testing.T) {
	tmpFilePublic, err := createTmpPublicKeyFile(tokentest.AlgoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	require.NoError(t, err)

	tmpFilePrivate, err := createTmpPrivateKeyFile(tokentest.AlgoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	require.NoError(t, err)

	expectedPrivateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
	require.NoError(t, err)
	expectedPublicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKeyFile", tmpFilePrivate.Name())
	config.Set("PublicKeyFile", tmpFilePublic.Name())
	config.Set("PlatformName", "my platform")
	tokenConfig, err := BuildConfig(config)
	require.NoError(t, err)
	assert.Equal(t, &Config{
		PrivateKey:   expectedPrivateKey,
		PublicKey:    expectedPublicKey,
		PlatformName: "my platform",
	}, tokenConfig)
}

func TestBuildConfig_LoadsKeysFromString(t *testing.T) {
	expectedPrivateKey, err := crypto.ParseRSAPrivateKeyFromPEM(tokentest.AlgoreaPlatformPrivateKey)
	require.NoError(t, err)
	expectedPublicKey, err := crypto.ParseRSAPublicKeyFromPEM(tokentest.AlgoreaPlatformPublicKey)
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKey", tokentest.AlgoreaPlatformPrivateKey)
	config.Set("PublicKey", tokentest.AlgoreaPlatformPublicKey)
	config.Set("PlatformName", "my platform")
	tokenConfig, err := BuildConfig(config)
	require.NoError(t, err)
	assert.Equal(t, &Config{
		PrivateKey:   expectedPrivateKey,
		PublicKey:    expectedPublicKey,
		PlatformName: "my platform",
	}, tokenConfig)
}

func TestBuildConfig_CannotLoadPublicKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile(tokentest.AlgoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKeyFile", tmpFilePrivate.Name())
	config.Set("PublicKeyFile", "nosuchfile.pem")
	config.Set("PlatformName", "my platform")
	_, err = BuildConfig(config)
	assert.IsType(t, &os.PathError{}, err)
}

func TestBuildConfig_CannotLoadPrivateKey(t *testing.T) {
	tmpFilePublic, err := createTmpPublicKeyFile(tokentest.AlgoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKeyFile", "nosuchfile.pem")
	config.Set("PublicKeyFile", tmpFilePublic.Name())
	config.Set("PlatformName", "my platform")
	_, err = BuildConfig(config)

	assert.IsType(t, &os.PathError{}, err)
}

func TestBuildConfig_CannotParsePublicKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile(tokentest.AlgoreaPlatformPrivateKey)
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	require.NoError(t, err)

	tmpFilePublic, err := createTmpPublicKeyFile([]byte{})
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKeyFile", tmpFilePrivate.Name())
	config.Set("PublicKeyFile", tmpFilePublic.Name())
	config.Set("PlatformName", "my platform")
	_, err = BuildConfig(config)

	assert.Equal(t, errors.New("invalid key: Key must be PEM encoded PKCS1 or PKCS8 private key"), err)
}

func TestBuildConfig_CannotParsePrivateKey(t *testing.T) {
	tmpFilePrivate, err := createTmpPrivateKeyFile([]byte{})
	if tmpFilePrivate != nil {
		defer func() { _ = os.Remove(tmpFilePrivate.Name()) }()
	}
	require.NoError(t, err)

	tmpFilePublic, err := createTmpPublicKeyFile(tokentest.AlgoreaPlatformPublicKey)
	if tmpFilePublic != nil {
		defer func() { _ = os.Remove(tmpFilePublic.Name()) }()
	}
	require.NoError(t, err)

	config := viper.New()
	config.Set("PrivateKeyFile", tmpFilePrivate.Name())
	config.Set("PublicKeyFile", tmpFilePublic.Name())
	config.Set("PlatformName", "my platform")
	_, err = BuildConfig(config)

	assert.Equal(t, errors.New("invalid key: Key must be PEM encoded PKCS1 or PKCS8 private key"), err)
}

func TestBuildConfig_MissingPublicKey(t *testing.T) {
	config := viper.New()
	config.Set("PlatformName", "my platform")
	_, err := BuildConfig(config)
	assert.EqualError(t, err, "missing Public key in the token config (PublicKey or PublicKeyFile)")
}

func TestBuildConfig_MissingPrivateKey(t *testing.T) {
	config := viper.New()
	config.Set("PlatformName", "my platform")
	config.Set("PublicKey", tokentest.AlgoreaPlatformPublicKey)
	_, err := BuildConfig(config)
	assert.EqualError(t, err, "missing Private key in the token config (PrivateKey or PrivateKeyFile)")
}

const relFileName = "app/token/token_test.go"

func Test_prepareFileName(t *testing.T) {
	assert.Equal(t, "/", prepareFileName("/"))

	// absolute path
	assert.Equal(t, "/afile.key", prepareFileName("/afile.key"))

	// rel path
	preparedFileName := prepareFileName(relFileName)
	assert.Equal(t, relFileName, preparedFileName[len(preparedFileName)-len(relFileName):])
	assert.FileExists(t, preparedFileName)
}

func Test_prepareFileName_StripsOnlyTheLastOccurrenceOfApp(t *testing.T) {
	monkey.Patch(os.Getwd, func() (string, error) { return "/app/something/app/ab/app/token", nil })
	defer monkey.UnpatchAll()
	relFileName := "app/token/token_test.go"
	preparedFileName := prepareFileName(relFileName)
	assert.Equal(t, "/app/something/app/ab/"+relFileName, preparedFileName)
}

func createTmpKeyFile(key []byte, fileName string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		return nil, err
	}

	_, err = tmpFile.Write(key)
	if err != nil {
		_ = tmpFile.Close()
		return nil, err
	}

	return tmpFile, nil
}

func createTmpPublicKeyFile(key []byte) (*os.File, error) {
	return createTmpKeyFile(key, "testPublicKey.pem")
}

func createTmpPrivateKeyFile(key []byte) (*os.File, error) {
	return createTmpKeyFile(key, "testPrivateKey.pem")
}
