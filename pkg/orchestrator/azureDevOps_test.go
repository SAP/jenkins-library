package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzure(t *testing.T) {
	t.Run("Azure - BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")
		os.Setenv("BUILD_SOURCEBRANCH", "refs/heads/feat/test-azure")
		os.Setenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI", "https://pogo.foo")
		os.Setenv("SYSTEM_TEAMPROJECT", "bar")
		os.Setenv("BUILD_BUILDID", "42")
		os.Setenv("BUILD_SOURCEVERSION", "abcdef42713")
		os.Setenv("BUILD_REPOSITORY_URI", "github.com/foo/bar")

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", p.GetBranch())
		assert.Equal(t, "https://pogo.foobar/_build/results?buildId=42", p.GetBuildUrl())
		assert.Equal(t, "abcdef42713", p.GetCommit())
		assert.Equal(t, "github.com/foo/bar", p.GetRepoUrl())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
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

	t.Run("PR - Branch Policy", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "feat/test-azure")
		os.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "123456789")
		os.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER", "42")
		os.Setenv("BUILD_REASON", "PullRequest")

		p := AzureDevOpsConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.True(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-azure", c.Branch)
		assert.Equal(t, "main", c.Base)
		assert.Equal(t, "42", c.Key)
	})

	t.Run("Azure DevOps - false", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		os.Setenv("AZURE_HTTP_USER_AGENT", "false")

		o := DetectOrchestrator()

		assert.Equal(t, Orchestrator(Unknown), o)
	})
}
