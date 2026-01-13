package python

import "fmt"

func PublishPackage(
	executeFn func(executable string, params ...string) error,
	virtualEnv string,
	repository string,
	username string,
	password string,
) error {
	// install dependency
	if err := InstallTwine(executeFn, virtualEnv); err != nil {
		return fmt.Errorf("failed to install twine module: %w", err)
	}
	// publish project
	return executeFn(
		getBinary(virtualEnv, "twine"),
		"upload",
		"--username", username,
		"--password", password,
		"--repository-url", repository,
		"--disable-progress-bar",
		"dist/*",
	)
}
