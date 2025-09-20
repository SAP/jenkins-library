package python

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

func BuildWithSetupPy(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	pythonArgs []string,
	setupArgs []string,
) error {
	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, "setup.py")
	flags = append(flags, setupArgs...)
	flags = append(flags, "sdist", "bdist_wheel")

	pythonBinary := "python"
	if len(virtualEnv) > 0 {
		pythonBinary = filepath.Join(virtualEnv, "bin", pythonBinary)
	}

	log.Entry().Debug("building project")
	if err := executeFn(pythonBinary, flags...); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}
	return nil
}
