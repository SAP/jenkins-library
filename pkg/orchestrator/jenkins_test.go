package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJenkins(t *testing.T) {
	t.Run("BranchBuild", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "https://foo.bar/baz")
		os.Setenv("BRANCH_NAME", "feat/test-jenkins")

		p, _ := NewOrchestratorSpecificConfigProvider()
		c := p.GetBranchBuildConfig()

		assert.False(t, p.IsPullRequest())
		assert.Equal(t, "feat/test-jenkins", c.Branch)
	})

	t.Run("PR", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
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
