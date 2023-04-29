//go:build unit
// +build unit

package orchestrator

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnknownOrchestrator(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		p, _ := NewOrchestratorSpecificConfigProvider()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "n/a", p.GetBuildURL())
		assert.Equal(t, "n/a", p.GetBranch())
		assert.Equal(t, "n/a", p.GetCommit())
		assert.Equal(t, "n/a", p.GetRepoURL())
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

	t.Run("env variables", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		p := UnknownOrchestratorConfigProvider{}

		assert.Equal(t, "n/a", p.OrchestratorVersion())
		assert.Equal(t, "n/a", p.GetBuildID())
		assert.Equal(t, "n/a", p.GetJobName())
		assert.Equal(t, "Unknown", p.OrchestratorType())
		assert.Equal(t, time.Time{}.UTC(), p.GetPipelineStartTime())
		assert.Equal(t, "FAILURE", p.GetBuildStatus())
		assert.Equal(t, "n/a", p.GetRepoURL())
		assert.Equal(t, "n/a", p.GetBuildURL())
		assert.Equal(t, "n/a", p.GetStageName())
		log, err := p.GetLog()
		assert.Equal(t, []byte{}, log)
		assert.Equal(t, nil, err)
	})
}
