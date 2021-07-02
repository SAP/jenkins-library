package orchestrator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrchestrator(t *testing.T) {
	t.Run("Not running on CI", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		_, err := NewOrchestratorSpecificConfigProvider()

		assert.EqualError(t, err, "unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
	})

	t.Run("Test orchestrator.toString()", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")

		o := DetectOrchestrator()

		assert.Equal(t, "AzureDevOps", o.String())
	})

	t.Run("Test areIndicatingEnvVarsSet", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}

		os.Setenv("GITHUB_ACTION", "true")
		tmp := areIndicatingEnvVarsSet(envVars)
		assert.True(t, tmp)

		os.Unsetenv("GITHUB_ACTION")
		os.Setenv("GITHUB_ACTIONS", "true")
		tmp = areIndicatingEnvVarsSet(envVars)
		assert.True(t, tmp)

		os.Setenv("GITHUB_ACTION", "1")
		os.Setenv("GITHUB_ACTIONS", "false")
		tmp = areIndicatingEnvVarsSet(envVars)
		assert.True(t, tmp)

		os.Setenv("GITHUB_ACTION", "false")
		os.Setenv("GITHUB_ACTIONS", "0")
		tmp = areIndicatingEnvVarsSet(envVars)
		assert.False(t, tmp)
	})
}
