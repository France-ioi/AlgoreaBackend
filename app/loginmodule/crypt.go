package loginmodule

import "crypto/aes"

func decryptAes128Ecb(data, key []byte) []byte {
	cipher, err := aes.NewCipher(key)
	mustNotBeError(err)
	decrypted := make([]byte, len(data))
	size := 16

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	// see https://secure.php.net/manual/de/function.openssl-encrypt.php#109598
	if len(decrypted) > 0 {
		paddingCharacter := decrypted[len(decrypted)-1]
		stripPadding := true
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
