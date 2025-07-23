// Package token provides a way to serialize/un-serialize data structures in an encrypted token.
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

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
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	today := time.Now().UTC()
	const oneDay = 24 * time.Hour
	yesterday := today.Add(-oneDay)
	tomorrow := today.Add(oneDay)

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

// UnmarshalDependingOnItemPlatform unmarshals a token from JSON representation
// using a platform's public key for given itemID.
// The function returns nil (success) if the platform doesn't use tokens.
func UnmarshalDependingOnItemPlatform[P any](
	store *database.DataStore,
	itemID int64,
	target **Token[P],
	token []byte,
	tokenFieldName string,
) (platformHasKey bool, err error) {
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
		logging.EntryFromContext(store.GetContext()).
			Warnf("cannot parse platform's public key for item with id = %d: %s", itemID, err.Error())

		return true, fmt.Errorf("invalid %s: wrong platform's key", tokenFieldName)
	}

	*target = &Token[P]{PublicKey: parsedPublicKey}

	if err = (*target).UnmarshalJSON(token); err != nil {
		return true, fmt.Errorf("invalid %s: %s", tokenFieldName, err.Error())
	}

	return true, nil
}
