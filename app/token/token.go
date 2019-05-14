package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/jose.v1/crypto"
	"gopkg.in/jose.v1/jws"

	"github.com/France-ioi/AlgoreaBackend/app/config"
)

// Config contains parsed keys and PlatformName
type Config struct {
	PublicKey    *rsa.PublicKey
	PrivateKey   *rsa.PrivateKey
	PlatformName string
}

// Initialize loads keys from the config and resolves the platform name
func Initialize(conf *config.Token) (tokenConfig *Config, err error) {
	tokenConfig = &Config{PlatformName: conf.PlatformName}

	bytes, err := ioutil.ReadFile(prepareFileName(conf.PublicKeyFile))
	if err != nil {
		return
	}
	tokenConfig.PublicKey, err = crypto.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadFile(prepareFileName(conf.PrivateKeyFile))
	if err != nil {
		return
	}
	tokenConfig.PrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return
	}
	return
}

func prepareFileName(fileName string) string {
	if len(fileName) > 0 && fileName[0] == '/' {
		return fileName
	}

	_, codeFilePath, _, _ := runtime.Caller(0)
	codeDir := filepath.Dir(codeFilePath)
	return codeDir + "/../../" + fileName
}

// ParseAndValidate parses a token and validates its signature and date
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

// Generate generates a signed token for a payload
func Generate(payload map[string]interface{}, privateKey *rsa.PrivateKey) []byte {
	payload["date"] = time.Now().UTC().Format("02-01-2006")

	token, err := jws.NewJWT(payload, crypto.SigningMethodRS512).Serialize(privateKey)
	if err != nil {
		panic(err)
	}
	return token
}
