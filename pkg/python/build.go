package python

import (
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
)

func Build(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	binaryFlags []string,
	moduleFlags []string,
) error {
	pythonBinary := "python"
	if len(virtualEnv) > 0 {
		pythonBinary = filepath.Join(virtualEnv, "bin", pythonBinary)
	}

	var flags []string
	flags = append(flags, binaryFlags...)
	flags = append(flags, "-m", "build", "--no-isolation")
	flags = append(flags, moduleFlags...)

	log.Entry().Debug("building project")
	if err := executeFn(pythonBinary, flags...); err != nil {
		return err
	}
	return nil
}
