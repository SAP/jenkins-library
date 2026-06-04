package python

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	JUnitReportFile    = "TEST-python.xml"
	CoverageReportFile = "cobertura-coverage.xml"
)

func RunTests(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	testOptions []string,
	junitPath string,
	coveragePath string,
) error {
	log.Entry().Debug("running python tests")
	// Reject testOptions that would silently relocate the report files managed
	// by this step. The GCS upload globs in pythonBuild metadata are pinned to
	// JUnitReportFile and CoverageReportFile; overriding them via testOptions
	// causes reports to land at a different path while the upload glob matches
	// nothing — producing a silent green build with no artifacts in GCS.
	for _, opt := range testOptions {
		if strings.HasPrefix(opt, "--junitxml") {
			return fmt.Errorf("testOptions must not override --junitxml; the report path is managed by the step (got %q)", opt)
		}
		if strings.HasPrefix(opt, "--cov-report=xml") {
			return fmt.Errorf("testOptions must not override --cov-report=xml; the report path is managed by the step (got %q)", opt)
		}
	}
	args := []string{
		fmt.Sprintf("--junitxml=%s", junitPath),
		"--cov",
		fmt.Sprintf("--cov-report=xml:%s", coveragePath),
	}
	args = append(args, testOptions...)
	if err := executeFn(getBinary(virtualEnv, "pytest"), args...); err != nil {
		if strings.Contains(err.Error(), "exit status 5") {
			return fmt.Errorf("pytest collected no tests — ensure your project has tests under a discoverable path (default: ./tests): %w", err)
		}
		return fmt.Errorf("pytest execution failed: %w", err)
	}
	return nil
}
