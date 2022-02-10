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
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_SERVER_URL", "github.com/")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "github.com/foo/bar/actions/runs/42", p.GetBuildUrl())
		assert.Equal(t, "feat/test-gh-actions", p.GetBranch())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoUrl())
		assert.Equal(t, "GitHubActions", p.OrchestratorType())
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
