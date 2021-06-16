package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTravis(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
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
		os.Clearenv()
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
