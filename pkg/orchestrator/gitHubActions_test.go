package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGitHubActions(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Unsetenv("GITHUB_HEAD_REF")
		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF_NAME", "feat/test-gh-actions")
		os.Setenv("GITHUB_REF", "refs/heads/feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")

		p, _ := func() (OrchestratorSpecificConfigProviding, error) {
			return &GitHubActionsConfigProvider{
				run: run{
					HtmlUrl:      "https://github.com/foo/bar/actions/runs/42",
					RunStartedAt: time.Time{},
					HeadCommit: struct {
						Id        string    `json:"id"`
						Timestamp time.Time `json:"timestamp"`
					}{
						// to be filled
					},
					Repository: struct {
						HtmlUrl string `json:"html_url"`
					}{
						HtmlUrl: "https://github.com/foo/bar",
					},
				},
			}, nil
		}()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "https://github.com/foo/bar/actions/runs/42", p.GetBuildURL())
		assert.Equal(t, "feat/test-gh-actions", p.GetBranch())
		assert.Equal(t, "refs/heads/feat/test-gh-actions", p.GetReference())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "https://github.com/foo/bar", p.GetRepoURL())
		assert.Equal(t, "GitHub Actions", p.OrchestratorType())
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
