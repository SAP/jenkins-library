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

	t.Run("Azure DevOps - false", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		os.Setenv("AZURE_HTTP_USER_AGENT", "false")

		o, err := DetectOrchestrator()

		assert.EqualError(t, err, "unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins, Travis)")
		assert.Equal(t, Orchestrator(Unknown), o)
	})
}
