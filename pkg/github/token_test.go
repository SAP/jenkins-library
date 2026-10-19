//go:build unit

package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithActionsTokenFallback(t *testing.T) {
	t.Run("configured token wins over GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv(GitHubTokenEnv, "env-token")
		username, token := WithActionsTokenFallback("user", "configured")
		assert.Equal(t, "user", username)
		assert.Equal(t, "configured", token)
	})

	t.Run("GITHUB_TOKEN used as last resort, username defaulted", func(t *testing.T) {
		t.Setenv(GitHubTokenEnv, "env-token")
		username, token := WithActionsTokenFallback("", "")
		assert.Equal(t, "x-access-token", username)
		assert.Equal(t, "env-token", token)
	})

	t.Run("token without username gets default username", func(t *testing.T) {
		t.Setenv(GitHubTokenEnv, "")
		username, token := WithActionsTokenFallback("", "st-token")
		assert.Equal(t, "x-access-token", username)
		assert.Equal(t, "st-token", token)
	})

	t.Run("nothing available stays empty", func(t *testing.T) {
		t.Setenv(GitHubTokenEnv, "")
		username, token := WithActionsTokenFallback("", "")
		assert.Empty(t, username)
		assert.Empty(t, token)
	})
}
