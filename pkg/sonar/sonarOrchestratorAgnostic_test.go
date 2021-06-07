package sonar

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetEnv(e []string) {
	os.Clearenv()
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}

func TestAzure(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("BUILD_SOURCEBRANCH", "refs/heads/feat/test-azure")
		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")
		os.Setenv("BUILD_REASON", "pogo")

		p, _ := NewOrchestratorSpecificConfigProvider()
		c := p.GetBranchBuildConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", c.Branch)
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "feat/test-azure")
		os.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "42")
		os.Setenv("BUILD_REASON", "PullRequest")

		p := AzureDevOpsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("Not running on CI", func(t *testing.T) {
		defer os.Setenv("AZURE_HTTP_USER_AGENT", os.Getenv("AZURE_HTTP_USER_AGENT"))
		os.Unsetenv("AZURE_HTTP_USER_AGENT")

		_, err := NewOrchestratorSpecificConfigProvider()

		assert.EqualError(t, err, "could not detect orchestrator. Supported is: Azure DevOps, GitHub Actions, Travis, Jenkins")
	})
}
func TestGitHubActions(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
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
func TestJenkins(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("JENKINS_URL", "https://foo.bar/baz")
		os.Setenv("BRANCH_NAME", "feat/test-jenkins")

		p, _ := NewOrchestratorSpecificConfigProvider()
		c := p.GetBranchBuildConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-jenkins", c.Branch)
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("BRANCH_NAME", "PR-42")
		os.Setenv("CHANGE_BRANCH", "feat/test-jenkins")
		os.Setenv("CHANGE_TARGET", "main")
		os.Setenv("CHANGE_ID", "42")

		p := JenkinsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-jenkins", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})
}

func TestTravis(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("TRAVIS", "true")
		os.Setenv("TRAVIS_BRANCH", "feat/test-travis")
		os.Setenv("TRAVIS_PULL_REQUEST", "false")

		p, _ := NewOrchestratorSpecificConfigProvider()
		c := p.GetBranchBuildConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-travis", c.Branch)
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Setenv("TRAVIS_PULL_REQUEST_BRANCH", "feat/test-travis")
		os.Setenv("TRAVIS_BRANCH", "main")
		os.Setenv("TRAVIS_PULL_REQUEST", "42")

		p := TravisConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-travis", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})
}
