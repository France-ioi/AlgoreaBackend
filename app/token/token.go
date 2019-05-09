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

var (
	platformPublicKey  *rsa.PublicKey
	platformPrivateKey *rsa.PrivateKey
	platformName       string
)

// Initialize loads keys from the config and sets the platform name
func Initialize(conf *config.Root) error {
	platformName = conf.Platform.Name

	var err error
	bytes, err := ioutil.ReadFile(prepareFileName(conf.Platform.PublicKeyFile))
	if err != nil {
		return err
	}
	platformPublicKey, err = crypto.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return err
	}
	bytes, err = ioutil.ReadFile(prepareFileName(conf.Platform.PrivateKeyFile))
	if err != nil {
		return err
	}
	platformPrivateKey, err = crypto.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return err
	}
	return nil
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
func ParseAndValidate(token []byte) (map[string]interface{}, error) {
	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return nil, err
	}

	// Validate token
	if err = jwt.Validate(platformPublicKey, crypto.SigningMethodRS512); err != nil {
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
func Generate(payload map[string]interface{}) []byte {
	payload["date"] = time.Now().UTC().Format("02-01-2006")
	payload["platformName"] = platformName

	token, err := jws.NewJWT(payload, crypto.SigningMethodRS512).Serialize(platformPrivateKey)
	if err != nil {
		panic(err)
	}
	return token
}
