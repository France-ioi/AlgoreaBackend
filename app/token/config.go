package token

import (
	"crypto/rsa"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/SermoDigital/jose/crypto"
	"github.com/spf13/viper"
)

// Config contains parsed keys and PlatformName.
type Config struct {
	PublicKey    *rsa.PublicKey
	PrivateKey   *rsa.PrivateKey
	PlatformName string
}

// BuildConfig loads keys from the config and resolves the platform name.
func BuildConfig(config *viper.Viper) (tokenConfig *Config, err error) {
	tokenConfig = &Config{PlatformName: config.GetString("PlatformName")}

	bytes, err := getKey(config, "Public")
	if err != nil {
		return
	}
	tokenConfig.PublicKey, err = crypto.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return
	}
	bytes, err = getKey(config, "Private")
	if err != nil {
		return
	}
	tokenConfig.PrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return
	}
	return
}

// getKey returns either "<keyType>Key" if not empty or the content of "<keyType>KeyFile" otherwise
// keyType is either "Public" or "Private".
func getKey(config *viper.Viper, keyType string) ([]byte, error) {
	key := config.GetString(keyType + "Key")
	if key != "" {
		return []byte(key), nil
	}
	if config.GetString(keyType+"KeyFile") == "" {
		return nil, fmt.Errorf("missing %s key in the token config (%sKey or %sKeyFile)", keyType, keyType, keyType)
	}
	return os.ReadFile(prepareFileName(config.GetString(keyType + "KeyFile")))
}

var tokenPathTestRegexp = regexp.MustCompile(`.*([/\\]app(?:[/\\][a-z]+)*?)$`)

func prepareFileName(fileName string) string {
	if fileName != "" && fileName[0] == '/' {
		return fileName
	}

	cwd, _ := os.Getwd()
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		match := tokenPathTestRegexp.FindStringSubmatchIndex(cwd)
		if match != nil {
			cwd = cwd[:match[2]]
		}
	}
	return cwd + "/" + fileName
}
