// Package token provides a way to serialize/un-serialize data structures in an encrypted token.
package token

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// Config contains parsed keys and PlatformName.
type Config struct {
	PublicKey    *rsa.PublicKey
	PrivateKey   *rsa.PrivateKey
	PlatformName string
}

// Initialize loads keys from the config and resolves the platform name.
func Initialize(config *viper.Viper) (tokenConfig *Config, err error) {
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
	if len(fileName) > 0 && fileName[0] == '/' {
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

// GetUnsafeFromToken returns the value of the field without checking the token signature.
func GetUnsafeFromToken(token []byte, field string) (interface{}, error) {
	token = []byte(strings.Trim(string(token), "\""))

	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return nil, err
	}
	return jwt.Claims().Get(field), nil
}

// ParseAndValidate parses a token and validates its signature and date.
func ParseAndValidate(token []byte, publicKey *rsa.PublicKey) (map[string]interface{}, error) {
	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return nil, err
	}

	// Validate token
	if err = jwt.Validate(publicKey, crypto.SigningMethodRS512); err != nil {
		return nil, fmt.Errorf("invalid token: %s", err)
	}

	today := time.Now().UTC()
	yesterday := today.Add(-24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	const dateLayout = "02-01-2006" // 'd-m-Y' in PHP
	todayStr := today.Format(dateLayout)
	yesterdayStr := yesterday.Format(dateLayout)
	tomorrowStr := tomorrow.Format(dateLayout)

	jwtDate := jwt.Claims().Get("date")
	if jwtDate != yesterdayStr && jwtDate != todayStr && jwtDate != tomorrowStr {
		return nil, errors.New("the token has expired")
	}

	return jwt.Claims(), nil
}

// Generate generates a signed token for a payload.
func Generate(payload map[string]interface{}, privateKey *rsa.PrivateKey) []byte {
	payload["date"] = time.Now().UTC().Format("02-01-2006")

	token, err := jws.NewJWT(payload, crypto.SigningMethodRS512).Serialize(privateKey)
	if err != nil {
		panic(err)
	}
	return token
}

// UnexpectedError represents an unexpected error so that we could differentiate it from expected errors.
type UnexpectedError struct {
	err error
}

// Error returns a string representation for an unexpected error.
func (ue *UnexpectedError) Error() string {
	return ue.err.Error()
}

// IsUnexpectedError returns true if its argument is an unexpected error.
func IsUnexpectedError(err error) bool {
	if _, unexpected := err.(*UnexpectedError); unexpected {
		return true
	}
	return false
}

// UnmarshalDependingOnItemPlatform unmarshals a token from JSON representation
// using a platform's public key for given itemID.
// The function returns nil (success) if the platform doesn't use tokens.
func UnmarshalDependingOnItemPlatform(
	store *database.DataStore,
	itemID int64,
	target interface{},
	token []byte,
	tokenFieldName string,
) (platformHasKey bool, err error) {
	targetRefl := reflect.ValueOf(target)
	defer recoverPanics(&err)

	publicKey, err := store.Platforms().GetPublicKeyByItemID(itemID)
	if gorm.IsRecordNotFoundError(err) {
		return false, fmt.Errorf("cannot find the platform for item %d", itemID)
	}
	mustNotBeError(err)

	if publicKey == nil {
		return false, nil
	}

	// Token shouldn't be null when there is a public key.
	if token == nil {
		return true, fmt.Errorf("missing %s", tokenFieldName)
	}

	parsedPublicKey, err := crypto.ParseRSAPublicKeyFromPEM([]byte(*publicKey))
	if err != nil {
		logging.SharedLogger.WithContext(store.GetContext()).
			Warnf("cannot parse platform's public key for item with id = %d: %s", itemID, err.Error())

		return true, fmt.Errorf("invalid %s: wrong platform's key", tokenFieldName)
	}

	targetRefl.Elem().Set(reflect.New(targetRefl.Elem().Type().Elem()))
	targetRefl.Elem().Elem().FieldByName("PublicKey").Set(reflect.ValueOf(parsedPublicKey))

	if err = targetRefl.Elem().Interface().(json.Unmarshaler).UnmarshalJSON(token); err != nil {
		return true, fmt.Errorf("invalid %s: %s", tokenFieldName, err.Error())
	}

	return true, nil
}

func mustNotBeError(err error) {
	if err != nil {
		panic(err)
	}
}

func recoverPanics(
	err *error, //nolint:gocritic // we need the pointer as we replace the error with a panic
) {
	if r := recover(); r != nil {
		*err = &UnexpectedError{err: r.(error)}
	}
}
