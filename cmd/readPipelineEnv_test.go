package cmd

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCpeEncryption(t *testing.T) {
	secret := []byte("testKey!")
	payload := []byte(strings.Repeat("testString", 100))

	encrypted, err := encrypt(secret, payload)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	decrypted, err := decrypt(secret, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, decrypted, payload)
}
