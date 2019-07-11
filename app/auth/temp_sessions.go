package auth

import (
	crand "crypto/rand"
	"math/big"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// TemporaryUserSessionLifetimeInSeconds specifies the lifetime of the access token for a temporary user
const TemporaryUserSessionLifetimeInSeconds = int32(2 * time.Hour / time.Second) // 2 hours (7200 seconds)

// CreateNewTempSession creates a new session for a temporary user
func CreateNewTempSession(s *database.SessionStore, userID int64) (accessToken string, expiresIn int32, err error) {
	expiresIn = TemporaryUserSessionLifetimeInSeconds

	err = s.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
		accessToken, err = GenerateTempAccessToken()
		if err != nil {
			return err
		}
		return retryStore.Sessions().InsertMap(map[string]interface{}{
			"sAccessToken":    accessToken,
			"sExpirationDate": gorm.Expr("NOW() + INTERVAL ? SECOND", expiresIn),
			"idUser":          userID,
			"sIssuer":         "backend",
		})
	})
	if err != nil {
		accessToken = ""
		return
	}

	return
}

// GenerateTempAccessToken generate a random access token for a temporary user's session.
// Entropy of the generated token (assuming "crypto/rand" is well implemented) is 36^32, so ~165 bits.
func GenerateTempAccessToken() (string, error) {
	const allowedCharacters = "0123456789abcdefghijklmnopqrstuvwxyz"
	const allowedCharactersLength = len(allowedCharacters)
	const tokenLength = 32

	result := make([]byte, 0, tokenLength)
	for i := 0; i < tokenLength; i++ {
		index, err := crand.Int(crand.Reader, big.NewInt(int64(allowedCharactersLength)))
		if err != nil {
			return "", err
		}
		result = append(result, allowedCharacters[index.Int64()])
	}
	return string(result), nil
}
