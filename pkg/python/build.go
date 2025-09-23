package python

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"path/filepath"
)

func BuildWithSetupPy(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	pythonArgs []string,
	setupArgs []string,
) error {
	// install dependency
	if err := InstallWheel(executeFn, virtualEnv); err != nil {
		return err
	}

	pythonBinary := "python"
	if len(virtualEnv) > 0 {
		pythonBinary = filepath.Join(virtualEnv, "bin", pythonBinary)
	}

	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, setupArgs...)
	flags = append(flags, "sdist", "bdist_wheel")

	containsCustomBuild := false
	for _, flag := range flags {
		if flag == "pypa/build" || flag == "pypa/installer" {
			containsCustomBuild = true
			break
		}
	}

	if !containsCustomBuild {
		flags = append(flags, "setup.py")
	}

	log.Entry().Debug("building project")
	if err := executeFn(pythonBinary, flags...); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}
	return nil
}
