package python

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

var (
	PipInstallFlags = []string{"install", "--upgrade", "--root-user-action=ignore"}
)

func install(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	module string,
	version string,
	extraArgs []string,
) error {
	flags := PipInstallFlags
	if len(extraArgs) > 0 {
		flags = append(flags, extraArgs...)
	}
	if len(version) > 0 {
		module = fmt.Sprintf("%s==%s", module, version)
	}
	if len(module) > 0 {
		flags = append(flags, module)
	}

	return executeFn(getBinary(virtualEnv, "pip"), flags...)
}

func InstallPip(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("updating pip")
	return install(executeFn, virtualEnv, "pip", "", nil)
}

func InstallProjectDependencies(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing project dependencies")
	return install(executeFn, virtualEnv, ".", "", nil)
}

func InstallRequirements(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	requirementsFile string,
) error {
	log.Entry().Debug("installing requirements")
	return install(executeFn, virtualEnv, "", "", []string{"--requirement", requirementsFile})
}

func InstallBuild(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing build")
	return install(executeFn, virtualEnv, "build", "", nil)
}

func InstallWheel(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing wheel")
	return install(executeFn, virtualEnv, "wheel", "", nil)
}

func InstallTwine(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
) error {
	log.Entry().Debug("installing twine")
	return install(executeFn, virtualEnv, "twine", "", nil)
}

func InstallCycloneDX(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	cycloneDXVersion string,
) error {
	log.Entry().Debug("installing cyclonedx-bom")
	return install(executeFn, virtualEnv, "cyclonedx-bom", cycloneDXVersion, nil)
}
