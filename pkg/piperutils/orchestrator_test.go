package piperutils

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetEnv(e []string) {
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}

func TestOrchestrator(t *testing.T) {
	t.Run("No orchestrator set", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		o, err := DetectOrchestrator()

		assert.EqualError(t, err, "could not detect orchestrator. Supported is: Azure DevOps, GitHub Actions, Travis, Jenkins")
		assert.Equal(t, Orchestrator(-2), o)
	})

	t.Run("Azure DevOps", func(t *testing.T) {
		defer os.Unsetenv("AZURE_HTTP_USER_AGENT")

		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")

		o, err := DetectOrchestrator()

		assert.Nil(t, err)
		assert.Equal(t, AzureDevOps, o)
		assert.Equal(t, "AzureDevOps", o.String())
	})

	t.Run("Azure DevOps - false", func(t *testing.T) {
		defer os.Unsetenv("AZURE_HTTP_USER_AGENT")

		os.Setenv("AZURE_HTTP_USER_AGENT", "false")

		o, err := DetectOrchestrator()

		assert.EqualError(t, err, "could not detect orchestrator. Supported is: Azure DevOps, GitHub Actions, Travis, Jenkins")
		assert.Equal(t, Orchestrator(-2), o)
	})

	t.Run("GitHub Actions", func(t *testing.T) {
		defer os.Unsetenv("GITHUB_ACTIONS")

		os.Setenv("GITHUB_ACTIONS", "true")

		o, err := DetectOrchestrator()

		assert.Nil(t, err)
		assert.Equal(t, GitHubActions, o)
	})

	t.Run("Jenkins", func(t *testing.T) {
		defer os.Unsetenv("JENKINS_HOME")

		os.Setenv("JENKINS_URL", "https://foo.bar/baz")

		o, err := DetectOrchestrator()

		assert.Nil(t, err)
		assert.Equal(t, Jenkins, o)
	})

	t.Run("Travis", func(t *testing.T) {
		defer os.Unsetenv("TRAVIS")

		os.Setenv("TRAVIS", "true")

		o, err := DetectOrchestrator()

		assert.Nil(t, err)
		assert.Equal(t, Travis, o)
	})
}
