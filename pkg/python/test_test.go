//go:build unit
// +build unit

package python

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

// exitError returns a real *exec.ExitError with the given exit code by running
// a short-lived subprocess. This is necessary because errors.As cannot match
// a plain fmt.Errorf against *exec.ExitError.
func exitError(t *testing.T, code int) error {
	t.Helper()
	cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", code))
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit, got nil for code %d", code)
	}
	return err
}

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
	t.Parallel()
	tests := []struct {
		name        string
		virtualEnv  string
		testOptions []string
		execErr     error
		wantExec    string
		wantParams  []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "happy path - no extra options",
			testOptions: nil,
			wantExec:    "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
			},
		},
		{
			name:       "with virtualenv - uses venv pytest binary",
			virtualEnv: ".venv",
			wantExec:   filepath.Join(".venv", "bin", "pytest"),
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
			},
		},
		{
			name:        "happy path - with test options appended",
			testOptions: []string{"-v", "--tb=short"},
			wantExec:    "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"-v",
				"--tb=short",
			},
		},
		{
			name:        "conflicting --junitxml in testOptions is rejected",
			testOptions: []string{"--junitxml=my-results.xml"},
			wantErr:     true,
			errContains: "--junitxml",
		},
		{
			name:        "conflicting --junitxml= (equals form) in testOptions is rejected",
			testOptions: []string{"--junitxml="},
			wantErr:     true,
			errContains: "--junitxml",
		},
		{
			name:        "conflicting --junit-xml (hyphenated form) in testOptions is rejected",
			testOptions: []string{"--junit-xml=my-results.xml"},
			wantErr:     true,
			errContains: "--junit-xml",
		},
		{
			name:        "conflicting --cov-report=xml in testOptions is rejected",
			testOptions: []string{"--cov-report=xml:other.xml"},
			wantErr:     true,
			errContains: "--cov-report=xml",
		},
		{
			name:        "benign --cov-report=html passthrough is allowed",
			testOptions: []string{"--cov-report=html:htmlcov"},
			wantExec:    "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"--cov-report=html:htmlcov",
			},
		},
		{
			name:        "conflicting --cov-report xml:path (space-separated) is rejected",
			testOptions: []string{"--cov-report", "xml:other.xml"},
			wantErr:     true,
			errContains: "--cov-report",
		},
		{
			name:        "conflicting --cov-report xml (bare, no path) is rejected",
			testOptions: []string{"--cov-report", "xml"},
			wantErr:     true,
			errContains: "--cov-report",
		},
		{
			name:        "benign --cov-report term (space-separated, non-xml) is allowed",
			testOptions: []string{"--cov-report", "term"},
			wantExec:    "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"--cov-report",
				"term",
			},
		},
		{
			name:        "trailing bare --cov-report (no following element) is allowed through",
			testOptions: []string{"--cov-report"},
			wantExec:    "pytest",
			wantParams: []string{
				"--junitxml=TEST-python.xml",
				"--cov",
				"--cov-report=xml:cobertura-coverage.xml",
				"--cov-report",
			},
		},
		{
			name:        "benign -v and --tb=short passthrough is allowed",
			testOptions: []string{"-v", "--tb=short"},
			wantExec:    "pytest",
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
			t.Parallel()
			mockRunner := mock.ExecMockRunner{}
			if tt.execErr != nil {
				mockRunner.ShouldFailOnCommand = map[string]error{"pytest": tt.execErr}
			}

			err := RunTests(mockRunner.RunExecutable, tt.virtualEnv, tt.testOptions)

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

func TestRunTestsNonZeroExit(t *testing.T) {
	t.Parallel()
	mockRunner := mock.ExecMockRunner{}
	mockRunner.ShouldFailOnCommand = map[string]error{"pytest": exitError(t, 1)}

	err := RunTests(mockRunner.RunExecutable, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pytest")
}

func TestRunTestsExitCode5(t *testing.T) {
	t.Parallel()
	mockRunner := mock.ExecMockRunner{}
	mockRunner.ShouldFailOnCommand = map[string]error{"pytest": exitError(t, 5)}

	err := RunTests(mockRunner.RunExecutable, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pytest collected no tests")
}
