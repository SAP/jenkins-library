package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_envVarsAreSet(t *testing.T) {
	t.Run("Test envVarsAreSet", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()

		envVars := []string{"GITHUB_ACTION", "GITHUB_ACTIONS"}

		os.Setenv("GITHUB_ACTION", "true")
		tmp := envVarsAreSet(envVars)
		assert.True(t, tmp)

		os.Unsetenv("GITHUB_ACTION")
		os.Setenv("GITHUB_ACTIONS", "true")
		tmp = envVarsAreSet(envVars)
		assert.True(t, tmp)

		os.Setenv("GITHUB_ACTION", "1")
		os.Setenv("GITHUB_ACTIONS", "false")
		tmp = envVarsAreSet(envVars)
		assert.True(t, tmp)

		os.Setenv("GITHUB_ACTION", "false")
		os.Setenv("GITHUB_ACTIONS", "0")
		tmp = envVarsAreSet(envVars)
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
