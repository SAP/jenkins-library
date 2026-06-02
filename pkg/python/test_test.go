//go:build unit
// +build unit

package python

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunTests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		testOptions  []string
		junitPath    string
		coveragePath string
		execErr      error
		wantExec     string
		wantParams   []string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "happy path - no extra options",
			testOptions:  nil,
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantExec:     "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
			},
		},
		{
			name:         "happy path - with test options appended",
			testOptions:  []string{"-v", "--tb=short"},
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantExec:     "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"-v",
				"--tb=short",
			},
		},
		{
			name:         "custom paths appear verbatim in flags",
			testOptions:  nil,
			junitPath:    "custom/junit.xml",
			coveragePath: "reports/cov.xml",
			wantExec:     "pytest",
			wantParams: []string{
				"--junitxml=custom/junit.xml",
				"--cov",
				"--cov-report=xml:reports/cov.xml",
			},
		},
		{
			name:         "pytest non-zero exit returns wrapped error mentioning pytest",
			testOptions:  nil,
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			execErr:      fmt.Errorf("exit status 1"),
			wantErr:      true,
			errContains:  "pytest",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRunner := mock.ExecMockRunner{}
			if tt.execErr != nil {
				mockRunner.ShouldFailOnCommand = map[string]error{"pytest": tt.execErr}
			}

			err := RunTests(mockRunner.RunExecutable, "", tt.testOptions, tt.junitPath, tt.coveragePath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, mockRunner.Calls, 1)
			assert.Equal(t, tt.wantExec, mockRunner.Calls[0].Exec)
			assert.Equal(t, tt.wantParams, mockRunner.Calls[0].Params)
		})
	}
}

func TestRunTestsWithVirtualEnv(t *testing.T) {
	t.Parallel()
	mockRunner := mock.ExecMockRunner{}

	err := RunTests(mockRunner.RunExecutable, ".venv", nil, "TEST-python.xml", "cov.xml")

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(".venv", "bin", "pytest"), mockRunner.Calls[0].Exec)
}
