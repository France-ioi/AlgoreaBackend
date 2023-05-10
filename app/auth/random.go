package auth

import (
	crand "crypto/rand"
	"math/big"
)

// GenerateKey generate a random string that can be used as an access token for a temporary user's session.
// The entropy of the generated string (assuming "crypto/rand" is well implemented) is 36^32, so ~165 bits.
func GenerateKey() (string, error) {
	const allowedCharacters = "0123456789abcdefghijklmnopqrstuvwxyz"
	const allowedCharactersLength = len(allowedCharacters)
	const stringLength = 32

	result := make([]byte, 0, stringLength)
	for i := 0; i < stringLength; i++ {
		index, err := crand.Int(crand.Reader, big.NewInt(int64(allowedCharactersLength)))
		if err != nil {
			return "", err
		}
		result = append(result, allowedCharacters[index.Int64()])
	}
	return string(result), nil
}
