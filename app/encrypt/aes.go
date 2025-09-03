// Package encrypt provides utilities to encrypt and decrypt data.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// AES256GCM encrypts the plaintext using AES-256-GCM with the provided key.
// It returns the ciphertext with the nonce prepended.
func AES256GCM(key, plaintext []byte) []byte {
	block, err := aes.NewCipher(key)
	service.MustNotBeError(err)

	gcm, err := cipher.NewGCM(block)
	service.MustNotBeError(err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	service.MustNotBeError(err)

	// We put nonce as destination (first argument) to prepend it to the ciphertext.
	return gcm.Seal(nonce, nonce, plaintext, nil)
}

// DecryptAES256GCM decrypts the ciphertext using AES-256-GCM with the provided key.
func DecryptAES256GCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	service.MustNotBeError(err)

	gcm, err := cipher.NewGCM(block)
	service.MustNotBeError(err)

	// The nonce is at the beginning.
	nonce := ciphertext[0:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
