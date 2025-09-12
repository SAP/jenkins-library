package python

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	Binary = "python"
)

var (
	PipInstallFlags = []string{"install", "--upgrade", "--root-user-action=ignore"}
)

func Install(
	binary string,
	executeFn func(executable string, params ...string) error,
	module string,
	version string,
) error {
	// flags := append([]string{"-m", "pip"}, PipInstallFlags...)
	flags := PipInstallFlags

	if len(version) > 0 {
		module = fmt.Sprintf("%s==%s", module, version)
	}
	flags = append(flags, module)

	if err := executeFn(binary, flags...); err != nil {
		return fmt.Errorf("failed to install %s: %w", module, err)
	}
	return nil
}

func InstallProjectDependencies(
	binary string,
	executeFn func(executable string, params ...string) error,
) error {
	log.Entry().Debug("installing project dependencies")
	return Install(binary, executeFn, ".", "")
}

func InstallBuild(
	binary string,
	executeFn func(executable string, params ...string) error,
) error {
	log.Entry().Debug("installing build")
	return Install(binary, executeFn, "build", "")
}

func InstallPip(
	binary string,
	executeFn func(executable string, params ...string) error,
) error {
	log.Entry().Debug("updating pip")
	return Install(binary, executeFn, "pip", "")
}
