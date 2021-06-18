package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubActions(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF", "main")
		os.Unsetenv("GITHUB_HEAD_REF")

		p, _ := NewOrchestratorSpecificConfigProvider()
		c := p.GetBranchBuildConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "main", c.Branch)
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("GITHUB_HEAD_REF", "feat/test-gh-actions")
		os.Setenv("GITHUB_BASE_REF", "main")
		os.Setenv("GITHUB_EVENT_PULL_REQUEST_NUMBER", "42")

		p := GitHubActionsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-gh-actions", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})
}
