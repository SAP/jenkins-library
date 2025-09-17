package python

import "github.com/SAP/jenkins-library/pkg/log"

func Build(
	binary string,
	executeFn func(executable string, params ...string) error,
	binaryFlags []string,
	moduleFlags []string,
) error {
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
