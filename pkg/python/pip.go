package python

import (
	"fmt"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	Binary = "python"
)

var (
	PipInstallFlags = []string{"install", "--upgrade"}
)

func Install(
	executeFn func(executable string, params ...string) error,
	module string,
	version string,
	virtualEnvName string,
	virutalEnvironmentPathMap map[string]string,
) error {
	log.Entry().Debugf("installing  %s dependency", module)
	flags := []string{"-m", "pip", "install", "--upgrade"}
	if version == "" {
		flags = append(flags, module)
	} else {
		flags = append(flags, fmt.Sprintf("%s==%s", module, version))
	}

	if err := executeFn(virutalEnvironmentPathMap["pip"], flags...); err != nil {
		return err
	}
	virutalEnvironmentPathMap[module] = filepath.Join(virtualEnvName, "bin", module)
	return nil
}

func InstallProjectDependencies(
	executeFn func(executable string, params ...string) error,
	binary string,
) error {
	log.Entry().Debug("installing project dependencies")
	if err := executeFn(binary, "-m", "pip", "install", "."); err != nil {
		return err
	}
	return nil
}
