package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

func Decrypt(secretKey, hexCipher string) string {
	ciphertext, err := hex.DecodeString(hexCipher)
	if err != nil {
		panic(err)
	}

	// Create a new AES cipher using the key.
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		panic(err)
	}

	// Create a GCM cipher mode instance.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		panic("ciphertext too short")
	}

	// Extract nonce from ciphertext.
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext.
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	return string(plaintext)
}
