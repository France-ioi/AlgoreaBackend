package encrypt

import (
	"testing"
)

func Test_AES256GCMShouldProvideDifferentOutputForSameOutput(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("Hello, world!")

	ciphertext1, err := AES256GCM(key, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key, plaintext, err)
	}

	ciphertext2, err := AES256GCM(key, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key, plaintext, err)
	}

	if string(ciphertext1) == string(ciphertext2) {
		t.Errorf("AES256GCM(%v, %v) returned the same ciphertext for different calls", key, plaintext)
	}
}

func Test_DecryptAES256GCMShouldReturnOriginalPlaintext(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("Hello, world!")

	ciphertext, err := AES256GCM(key, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key, plaintext, err)
	}

	decrypted, err := DecryptAES256GCM(key, ciphertext)
	if err != nil {
		t.Errorf("DecryptAES256GCM(%v, %v) returned error: %v", key, ciphertext, err)
	}

	if string(plaintext) != string(decrypted) {
		t.Errorf("DecryptAES256GCM(%v, %v) returned %v, want %v", key, ciphertext, decrypted, plaintext)
	}
}

func Test_DecryptAES256GCMShouldReturnErrorForDifferentKey(t *testing.T) {
	key1 := []byte("12345678901234567890123456789012")
	key2 := []byte("12345678901234567890123456789013")
	plaintext := []byte("Hello, world!")

	ciphertext, err := AES256GCM(key1, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key1, plaintext, err)
	}

	_, err = DecryptAES256GCM(key2, ciphertext)
	if err == nil {
		t.Errorf("DecryptAES256GCM(%v, %v) returned nil error, want error", key2, ciphertext)
	}
}

func Test_DecryptAES256GCMShouldReturnErrorForDifferentNonce(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("Hello, world!")

	ciphertext, err := AES256GCM(key, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key, plaintext, err)
	}

	// The nonce is at the beginning.
	ciphertext[0] = ciphertext[0] + 1

	_, err = DecryptAES256GCM(key, ciphertext)
	if err == nil {
		t.Errorf("DecryptAES256GCM(%v, %v) returned nil error, want error", key, ciphertext)
	}
}

func Test_DecryptAES256GCMShouldReturnErrorForDifferentCiphertext(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("Hello, world!")

	ciphertext, err := AES256GCM(key, plaintext)
	if err != nil {
		t.Errorf("AES256GCM(%v, %v) returned error: %v", key, plaintext, err)
	}

	// The ciphertext is after the nonce, so the last byte is necessarily the ciphertext.
	ciphertext[len(ciphertext)-1] = ciphertext[1] + 1

	_, err = DecryptAES256GCM(key, ciphertext)
	if err == nil {
		t.Errorf("DecryptAES256GCM(%v, %v) returned nil error, want error", key, ciphertext)
	}
}
