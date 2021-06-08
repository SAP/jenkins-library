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

		assert.EqualError(t, err, "could not detect orchestrator. Supported is: Azure DevOps, GitHub Actions, Travis, Jenkins")
	})

	t.Run("Test orchestrator.toString()", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")

		o, err := DetectOrchestrator()

		assert.Nil(t, err)
		assert.Equal(t, "AzureDevOps", o.String())
	})
}
