package python

import (
	"path/filepath"
)

func PublishPackage(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	repository string,
	username string,
	password string,
) error {
	// install dependency
	if err := InstallTwine(executeFn, virtualEnv); err != nil {
		return err
	}
	// handle virtual environment
	twineBinary := "twine"
	if len(virtualEnv) > 0 {
		twineBinary = filepath.Join(virtualEnv, "bin", twineBinary)
	}
	// publish project
	return executeFn(
		twineBinary,
		"upload",
		"--username", username,
		"--password", password,
		"--repository-url", repository,
		"--disable-progress-bar",
		"dist/*",
	)
}
