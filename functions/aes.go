// Module to handle all of the encryption

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"math/big"
)

// Generate the 256-bit key, if the secret is less than 256-bit add padding, if its bigger trim it
func gen_key(secret big.Int) []byte {
	secret_bytes := []byte(secret.String())

	key := []byte{}

	if len(secret_bytes) < 32 {
		key = append(key, secret_bytes...)
		key = append(key, make([]byte, 32-len((secret_bytes)))...)
	} else {
		key = []byte(secret.String())[:32]
	}
	return key
}

// Decrypt data that has been sent with AES-GCM (Galois/Counter Mode)
func decrypt(secret big.Int, encrypted []byte) []byte {
	key := gen_key(secret)

	block, _ := aes.NewCipher(key)

	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize()) // Empty nonce

	encrypted, _ = gcm.Open(nil, nonce, encrypted, nil)

	return encrypted
}

// Encrypt data that has been received with AES-GCM (Galois/Counter Mode)
func encrypt(secret big.Int, msg []byte) []byte {
	key := gen_key(secret)

	block, _ := aes.NewCipher(key)

	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize()) // Empty nonce

	return gcm.Seal(nil, nonce, msg, nil)
}
