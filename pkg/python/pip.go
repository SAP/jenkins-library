package python

import "path/filepath"

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
	var flags []string
	if version == "" {
		flags = append(PipInstallFlags, module)
	} else {
		flags = append(PipInstallFlags, module+"=="+version)
	}

	if err := executeFn(virutalEnvironmentPathMap["pip"], flags...); err != nil {
		return err
	}
	virutalEnvironmentPathMap[module] = filepath.Join(virtualEnvName, "bin", module)
	return nil
}
