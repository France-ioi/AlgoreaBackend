package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// AES256GCM encrypts the plaintext using AES-256-GCM with the provided key.
// It returns the ciphertext with the nonce prepended.
func AES256GCM(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// We put nonce as destination (first argument) to prepend it to the ciphertext.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES256GCM decrypts the ciphertext using AES-256-GCM with the provided key.
func DecryptAES256GCM(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// The nonce is at the beginning.
	nonce := ciphertext[0:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
