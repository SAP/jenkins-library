package python

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

var (
	PipInstallFlags = []string{"install", "--upgrade", "--root-user-action=ignore"}
)

func Install(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	module string,
	version string,
	extraArgs []string,
) error {
	pipBinary := "pip"
	if len(virtualEnv) > 0 {
		pipBinary = filepath.Join(virtualEnv, "bin", pipBinary)
	}

	flags := PipInstallFlags
	// flags := append([]string{"-m", "pip"}, PipInstallFlags...)
	if len(version) > 0 {
		module = fmt.Sprintf("%s==%s", module, version)
	}
	if len(module) > 0 {
		flags = append(flags, module)
	}

	if err := executeFn(pipBinary, flags...); err != nil {
		return fmt.Errorf("failed to install %s: %w", module, err)
	}
	return nil
}

func InstallPip(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("updating pip")
	return Install(executeFn, virtualEnv, "pip", "")
}

func InstallProjectDependencies(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing project dependencies")
	return Install(executeFn, virtualEnv, ".", "")
}

func InstallBuild(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing build")
	return Install(executeFn, virtualEnv, "build", "")
}

func InstallWheel(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing wheel")
	return Install(executeFn, virtualEnv, "wheel", "")
}

func InstallTwine(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing twine")
	return Install(executeFn, virtualEnv, "twine", "")
}

func InstallCycloneDX(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	cycloneDXVersion string,
) error {
	log.Entry().Debug("installing cyclonedx-bom")
	return Install(executeFn, virtualEnv, "cyclonedx-bom", cycloneDXVersion)
}

func InstallRequirements(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	requirementsFile string,
) error {
	log.Entry().Debug("installing requirements")
	return Install(executeFn, virtualEnv, "", "", []string{"--requirement", requirementsFile})
}

func InstallWheel(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing wheel")
	return Install(executeFn, virtualEnv, "wheel", "", nil)
}

func InstallTwine(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing twine")
	return Install(executeFn, virtualEnv, "twine", "", nil)
}

func InstallCycloneDX(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	cycloneDXVersion string,
) error {
	log.Entry().Debug("installing cyclonedx-bom")
	return Install(executeFn, virtualEnv, "cyclonedx-bom", cycloneDXVersion, nil)
}
