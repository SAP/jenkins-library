package npm

import (
	"fmt"
)

// PublishAllPackages executes npm or yarn Install for all package.json fileUtils defined in packageJSONFiles
func (exec *Execute) PublishAllPackages(packageJSONFiles []string, registry string) error {
	for _, packageJSON := range packageJSONFiles {
		fileExists, err := exec.Utils.FileExists(packageJSON)
		if err != nil {
			return fmt.Errorf("cannot check if '%s' exists: %w", packageJSON, err)
		}
		if !fileExists {
			return fmt.Errorf("package.json file '%s' not found: %w", packageJSON, err)
		}

		err = exec.publish(packageJSON, registry)
		if err != nil {
			return err
		}
	}
	return nil
}

// publish executes npm publish for package.json
func (exec *Execute) publish(packageJSON, registry string) error {
	execRunner := exec.Utils.GetExecRunner()

	// oldWorkingDirectory, err := exec.Utils.Getwd()
	// if err != nil {
	// 	return fmt.Errorf("failed to get current working directory before executing npm scripts: %w", err)
	// }

	// dir := filepath.Dir(packageJSON)
	// err = exec.Utils.Chdir(dir)
	// if err != nil {
	// 	return fmt.Errorf("failed to change into directory for executing script: %w", err)
	// }

	// err = exec.SetNpmRegistries()
	// if err != nil {
	// 	return err
	// }

	// packageLockExists, yarnLockExists, err := exec.checkIfLockFilesExist()
	// if err != nil {
	// 	return err
	// }

	// log.Entry().WithField("WorkingDirectory", dir).Info("Running Publish")
	// if packageLockExists {
	// 	err = execRunner.RunExecutable("npm", "ci")
	// 	if err != nil {
	// 		return err
	// 	}
	// } else if yarnLockExists {
	// 	err = execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// log.Entry().Warn("No package lock file found. " +
	// 	"It is recommended to create a `package-lock.json` file by running `npm Install` locally." +
	// 	" Add this file to your version control. " +
	// 	"By doing so, the builds of your application become more reliable.")
	err := execRunner.RunExecutable("npm", "publish", packageJSON, "--dry-run", "--registery", registry)
	if err != nil {
		return err
	}
	// }

	// err = exec.Utils.Chdir(oldWorkingDirectory)
	// if err != nil {
	// 	return fmt.Errorf("failed to change back into original directory: %w", err)
	// }
	return nil
}
