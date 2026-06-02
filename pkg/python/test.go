package python

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

func RunTests(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	testOptions []string,
	junitPath string,
	coveragePath string,
) error {
	log.Entry().Debug("running python tests")
	args := []string{
		fmt.Sprintf("--junitxml=%s", junitPath),
		"--cov",
		fmt.Sprintf("--cov-report=xml:%s", coveragePath),
	}
	args = append(args, testOptions...)
	if err := executeFn(getBinary(virtualEnv, "pytest"), args...); err != nil {
		return fmt.Errorf("pytest execution failed: %w", err)
	}
	return nil
}
