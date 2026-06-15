package python

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	JUnitReportFile    = "TEST-python.xml"
	CoverageReportFile = "cobertura-coverage.xml"
)

// RunTests runs pytest inside virtualEnv with the given extra testOptions.
// executeFn must propagate *exec.ExitError transparently (via %w) so that the
// exit-code-5 "no tests collected" branch can unwrap it with errors.As.
func RunTests(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	testOptions []string,
) error {
	log.Entry().Debug("running python tests")
	// Reject testOptions that would silently relocate the report files managed
	// by this step. The GCS upload globs in pythonBuild metadata are pinned to
	// JUnitReportFile and CoverageReportFile; overriding them via testOptions
	// causes reports to land at a different path while the upload glob matches
	// nothing — producing a silent green build with no artifacts in GCS.
	for i, opt := range testOptions {
		if strings.HasPrefix(opt, "--junitxml") || strings.HasPrefix(opt, "--junit-xml") {
			return fmt.Errorf("testOptions must not override --junitxml/--junit-xml; the report path is managed by the step (got %q)", opt)
		}
		// Equals-separated form: --cov-report=xml[:path]
		if strings.HasPrefix(opt, "--cov-report=xml") {
			return fmt.Errorf("testOptions must not override --cov-report=xml; the report path is managed by the step (got %q)", opt)
		}
		// Space-separated form: --cov-report xml[:path]
		if opt == "--cov-report" && i+1 < len(testOptions) && strings.HasPrefix(testOptions[i+1], "xml") {
			return fmt.Errorf("testOptions must not override --cov-report xml; the report path is managed by the step (got %q %q)", opt, testOptions[i+1])
		}
	}
	args := []string{
		"--junitxml=" + JUnitReportFile,
		"--cov",
		"--cov-report=xml:" + CoverageReportFile,
	}
	args = append(args, testOptions...)
	if err := executeFn(getBinary(virtualEnv, "pytest"), args...); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 5 {
			return fmt.Errorf("pytest collected no tests — ensure your project has tests under a discoverable path (default: ./tests): %w", err)
		}
		return fmt.Errorf("pytest execution failed: %w", err)
	}
	return nil
}
