package loginmodule

import (
	"crypto/aes"
)

const blockSize = 16

func decryptAes128Ecb(data, key []byte) []byte {
	cipher, err := aes.NewCipher(key)
	mustNotBeError(err)
	decrypted := make([]byte, len(data))

	for bs, be := 0, blockSize; be <= len(data); bs, be = bs+blockSize, be+blockSize {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	// see https://secure.php.net/manual/de/function.openssl-encrypt.php#109598
	if len(decrypted) > 0 {
		stripPadding := true

		paddingCharacter := decrypted[len(decrypted)-1]
		for i := 1; i <= int(paddingCharacter); i++ {
			if decrypted[len(decrypted)-i] != paddingCharacter {
				stripPadding = false
				break
			}
		}
		if stripPadding {
			decrypted = decrypted[0 : len(decrypted)-int(paddingCharacter)]
		}
	}

	return decrypted
}

func encryptAes128Ecb(data, key []byte) []byte {
	cipher, err := aes.NewCipher(key)
	mustNotBeError(err)

	paddingLength := blockSize - len(data)%blockSize
	dataCopy := make([]byte, 0, len(data)+paddingLength)
	dataCopy = append(dataCopy, data...)
	for i := 0; i < paddingLength; i++ {
		dataCopy = append(dataCopy, byte(paddingLength))
	}

	encrypted := make([]byte, len(dataCopy))

	for bs, be := 0, blockSize; bs < len(dataCopy); bs, be = bs+blockSize, be+blockSize {
		cipher.Encrypt(encrypted[bs:be], dataCopy[bs:be])
	}

	return encrypted
}
