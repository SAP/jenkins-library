package cmd

import (
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/encryption"
	"github.com/stretchr/testify/assert"
)

func TestCpeEncryption(t *testing.T) {
	secret := []byte("testKey!")
	payload := []byte(strings.Repeat("testString", 100))

	encrypted, err := encryption.Encrypt(secret, payload)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)

	decrypted, err := encryption.Decrypt(secret, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, decrypted, payload)
}
