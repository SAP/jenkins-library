package encryption

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecrypt(t *testing.T) {
	t.Run("successful decryption", func(t *testing.T) {
		secret := []byte("test-secret-key")
		plaintext := []byte("hello world")

		// Encrypt first using our package function
		encrypted, err := Encrypt(secret, plaintext)
		assert.NoError(t, err)

		// Test decryption
		decrypted, err := Decrypt(secret, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("invalid base64 input", func(t *testing.T) {
		secret := []byte("test-secret-key")
		invalidBase64 := []byte("this is not base64!")

		decrypted, err := Decrypt(secret, invalidBase64)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.Contains(t, err.Error(), "failed to decode from base64")
	})

	t.Run("input too small", func(t *testing.T) {
		secret := []byte("test-secret-key")
		tooSmall := base64.StdEncoding.EncodeToString([]byte("small"))

		decrypted, err := Decrypt(secret, []byte(tooSmall))
		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.Contains(t, err.Error(), "invalid ciphertext: block size too small")
	})

	t.Run("empty input", func(t *testing.T) {
		secret := []byte("test-secret-key")
		empty := []byte("")

		decrypted, err := Decrypt(secret, empty)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
	})
}

func TestEncrypt(t *testing.T) {
	t.Run("successful encryption", func(t *testing.T) {
		secret := []byte("test-secret-key")
		plaintext := []byte("hello world")

		result, err := Encrypt(secret, plaintext)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify we can decrypt it back
		decrypted, err := Decrypt(secret, result)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("empty input", func(t *testing.T) {
		secret := []byte("test-secret-key")
		empty := []byte("")

		result, err := Encrypt(secret, empty)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify we can decrypt it back
		decrypted, err := Decrypt(secret, result)
		assert.NoError(t, err)
		assert.Equal(t, empty, decrypted)
	})

	t.Run("empty secret", func(t *testing.T) {
		secret := []byte("")
		plaintext := []byte("hello world")

		result, err := Encrypt(secret, plaintext)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create cipher")
	})

	t.Run("large input", func(t *testing.T) {
		secret := []byte("test-secret-key")
		largeInput := []byte(strings.Repeat("large input test ", 1000))

		result, err := Encrypt(secret, largeInput)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify we can decrypt it back
		decrypted, err := Decrypt(secret, result)
		assert.NoError(t, err)
		assert.Equal(t, largeInput, decrypted)
	})
}
