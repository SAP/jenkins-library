package cmd

import (
	"os"
)

// Deprecated: Please use piperutils.Files{} instead
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

const golangBuildTool = "golang"

// prepare golang private packages for whitesource and blackduck(detectExecuteScan)
func prepareGolangPrivatePackages(privateModules, privateModulesGitToken string) error {

	goConfig := golangBuildOptions{
		PrivateModules:         privateModules,
		PrivateModulesGitToken: privateModulesGitToken,
	} // only these parameters are enough to configure

	utils := newGolangBuildUtils(goConfig)

	goModFile, err := readGoModFile(utils) // returns nil if go.mod doesnt exist
	if err != nil {
		return err
	}

	if err = prepareGolangEnvironment(&goConfig, goModFile, utils); err != nil {
		return err
	}

	return nil
}
