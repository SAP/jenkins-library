package python

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

func BuildWithSetupPy(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	pythonArgs []string,
	setupArgs []string,
) error {
	log.Entry().Debug("building project with setup.py")
	// install dependency
	if err := InstallWheel(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install wheel module: %w", err)
	}

	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, "setup.py")
	flags = append(flags, setupArgs...)
	flags = append(flags, "sdist", "bdist_wheel")

	log.Entry().Debugf("building project with flags: %v", flags)
	return executeFn(getBinary(virtualEnv, "python"), flags...)
}

func BuildWithPyProjectToml(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	pythonArgs []string,
	moduleArgs []string,
) error {
	log.Entry().Debug("building project with toml")
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

	var flags []string
	flags = append(flags, pythonArgs...)
	flags = append(flags, "-m", "build", "--no-isolation")
	flags = append(flags, moduleArgs...)

	log.Entry().Debugf("building project with flags: %v", flags)
	return executeFn(getBinary(virtualEnv, "python"), flags...)
}
