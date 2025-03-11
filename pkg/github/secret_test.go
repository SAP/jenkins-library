//go:build unit
// +build unit

package github

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestRunGithubCreateEncryptedSecret(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		mockKeyID := "1"
		mockB64Key := base64.StdEncoding.EncodeToString([]byte("testPublicKey"))
		mockPubKey := github.PublicKey{KeyID: &mockKeyID, Key: &mockB64Key}

		mockName := "testSecret"
		mockValue := "testValue"

		// test
		githubSecret, err := CreateEncryptedSecret(mockName, mockValue, &mockPubKey)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, mockName, githubSecret.Name)
		assert.Equal(t, mockKeyID, githubSecret.KeyID)
	})
}
