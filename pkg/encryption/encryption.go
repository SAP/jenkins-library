package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// Decrypt decrypts base64-encoded data using AES-CFB
func Decrypt(secret, base64CipherText []byte) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(string(base64CipherText))
	if err != nil {
		return nil, fmt.Errorf("failed to decode from base64: %w", err)
	}

	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(cipherText) < aes.BlockSize {
		return nil, fmt.Errorf("invalid ciphertext: block size too small")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}

// Encrypt encrypts data using AES-CFB and encodes it in base64
func Encrypt(secret, inBytes []byte) ([]byte, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("failed to create cipher: empty secret")
	}

	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(inBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to init iv: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], inBytes)

	return []byte(base64.StdEncoding.EncodeToString(cipherText)), nil
}
