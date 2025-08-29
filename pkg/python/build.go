package python

import "github.com/SAP/jenkins-library/pkg/log"

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

func Build(
	executeFn func(executable string, params ...string) error,
	binary string,
	binaryFlags []string,
	moduleFlags []string,
) error {
	// Set default value for binaryFlags if nil
	if binaryFlags == nil {
		binaryFlags = []string{}
	}

	var flags []string
	flags = append(flags, binaryFlags...)
	flags = append(flags, "-m", "build", "--no-isolation")
	flags = append(flags, moduleFlags...)

	log.Entry().Debug("building project")
	if err := executeFn(binary, flags...); err != nil {
		return err
	}
	return nil
}
