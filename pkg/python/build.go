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
	// install dependency
	if err := InstallWheel(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install wheel module: %w", err)
	}

	pythonBinary := "python"
	if len(virtualEnv) > 0 {
		pythonBinary = filepath.Join(virtualEnv, "bin", pythonBinary)
	}

	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, "setup.py")
	flags = append(flags, setupArgs...)
	flags = append(flags, "sdist", "bdist_wheel")

	log.Entry().Debug("building project")
	return executeFn(pythonBinary, flags...)
}

func BuildWithPyProjectToml(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	pythonArgs []string,
	moduleArgs []string,
) error {
	// install dependencies
	if err := InstallPip(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to upgrade pip: %w", err)
	}
	if err := InstallProjectDependencies(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install project dependencies: %w", err)
	}
	if err := InstallBuild(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install build module: %w", err)
	}
	if err := InstallWheel(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install wheel module: %w", err)
	}

	pythonBinary := "python"
	if len(virtualEnv) > 0 {
		pythonBinary = filepath.Join(virtualEnv, "bin", pythonBinary)
	}

	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, "-m", "build", "--no-isolation")
	flags = append(flags, moduleArgs...)

	log.Entry().Debug("building project")
	return executeFn(pythonBinary, flags...)
}
