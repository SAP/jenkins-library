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
	executeFn func(executable string, params ...string) error,
	pipBinary string,
	module string,
	version string,
) error {
	// flags := append([]string{"-m", "pip"}, PipInstallFlags...)
	flags := PipInstallFlags

	if len(version) > 0 {
		module = fmt.Sprintf("%s==%s", module, version)
	}
	flags = append(flags, module)

	if err := executeFn(pipBinary, flags...); err != nil {
		return fmt.Errorf("failed to install %s: %w", module, err)
	}
	return nil
}

func InstallProjectDependencies(
	executeFn func(executable string, params ...string) error,
	pipBinary string,
) error {
	log.Entry().Debug("installing project dependencies")
	return Install(executeFn, pipBinary, ".", "")
}

func InstallBuild(
	executeFn func(executable string, params ...string) error,
	pipBinary string,
) error {
	log.Entry().Debug("installing build")
	return Install(executeFn, pipBinary, "build", "")
}

func InstallWheel(
	executeFn func(executable string, params ...string) error,
	pipBinary string,
) error {
	log.Entry().Debug("installing wheel")
	return Install(executeFn, pipBinary, "wheel", "")
}

func InstallPip(
	executeFn func(executable string, params ...string) error,
	pipBinary string,
) error {
	log.Entry().Debug("updating pip")
	return Install(executeFn, pipBinary, "pip", "")
}
