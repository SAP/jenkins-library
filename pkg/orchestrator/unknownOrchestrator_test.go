package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownOrchestrator(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "n/a", p.GetBuildUrl())
		assert.Equal(t, "n/a", p.GetBranch())
		assert.Equal(t, "n/a", p.GetCommit())
		assert.Equal(t, "n/a", p.GetRepoUrl())
		assert.Equal(t, "Unknown", p.OrchestratorType())
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		p := UnknownOrchestratorConfigProvider{}
		c := p.GetPullRequestConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "n/a", c.Branch)
		assert.Equal(t, "n/a", c.Base)
		assert.Equal(t, "n/a", c.Key)
	})
}
