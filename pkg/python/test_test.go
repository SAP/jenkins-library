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

// TestReportFileConstants pins JUnitReportFile and CoverageReportFile to their
// expected string values. These constants are duplicated as glob patterns in
// resources/metadata/pythonBuild.yaml under the `reports` output resource
// ("**/TEST-python.xml" and "**/cobertura-coverage.xml"). If you change either
// constant you MUST also update the metadata YAML and re-run `go generate`.
func TestReportFileConstants(t *testing.T) {
	assert.Equal(t, "TEST-python.xml", JUnitReportFile,
		"JUnitReportFile must match the glob in resources/metadata/pythonBuild.yaml reports output")
	assert.Equal(t, "cobertura-coverage.xml", CoverageReportFile,
		"CoverageReportFile must match the glob in resources/metadata/pythonBuild.yaml reports output")
}

func TestRunTests(t *testing.T) {
	tests := []struct {
		name         string
		virtualEnv   string
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
			name:         "with virtualenv - uses venv pytest binary",
			virtualEnv:   ".venv",
			testOptions:  nil,
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantExec:     filepath.Join(".venv", "bin", "pytest"),
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
		{
			name:         "pytest exit status 5 (no tests collected) returns actionable message",
			testOptions:  nil,
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			execErr:      fmt.Errorf("exit status 5"),
			wantErr:      true,
			errContains:  "pytest collected no tests",
		},
		{
			name:         "conflicting --junitxml in testOptions is rejected",
			testOptions:  []string{"--junitxml=my-results.xml"},
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantErr:      true,
			errContains:  "--junitxml",
		},
		{
			name:         "conflicting --junitxml= (equals form) in testOptions is rejected",
			testOptions:  []string{"--junitxml="},
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantErr:      true,
			errContains:  "--junitxml",
		},
		{
			name:         "conflicting --cov-report=xml in testOptions is rejected",
			testOptions:  []string{"--cov-report=xml:other.xml"},
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantErr:      true,
			errContains:  "--cov-report=xml",
		},
		{
			name:         "benign --cov-report=html passthrough is allowed",
			testOptions:  []string{"--cov-report=html:htmlcov"},
			junitPath:    "TEST-python.xml",
			coveragePath: "cobertura-coverage.xml",
			wantExec:     "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"--cov-report=html:htmlcov",
			},
		},
		{
			name:         "benign -v and --tb=short passthrough is allowed",
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mock.ExecMockRunner{}
			if tt.execErr != nil {
				mockRunner.ShouldFailOnCommand = map[string]error{"pytest": tt.execErr}
			}

			err := RunTests(mockRunner.RunExecutable, tt.virtualEnv, tt.testOptions, tt.junitPath, tt.coveragePath)

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
