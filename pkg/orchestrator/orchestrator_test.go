//go:build unit
// +build unit

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

		provider, err := NewOrchestratorSpecificConfigProvider()

		assert.EqualError(t, err, "unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
		assert.Equal(t, "Unknown", provider.OrchestratorType())
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

func Test_getEnv(t *testing.T) {
	type args struct {
		key      string
		fallback string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		envVar string
	}{
		{
			name:   "environment variable found",
			args:   args{key: "debug", fallback: "fallback"},
			want:   "found",
			envVar: "debug",
		},
		{
			name: "fallback variable",
			args: args{key: "debug", fallback: "fallback"},
			want: "fallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv(os.Environ())
			os.Clearenv()
			os.Setenv(tt.envVar, "found")
			assert.Equalf(t, tt.want, getEnv(tt.args.key, tt.args.fallback), "getEnv(%v, %v)", tt.args.key, tt.args.fallback)
		})
	}
}
