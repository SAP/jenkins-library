package npm

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

const pnpmPath = tmpInstallFolder + "/node_modules/.bin/pnpm"

// PackageManager represents a Node.js package manager configuration
type PackageManager struct {
	Name           string
	LockFile       string
	InstallCommand string
	InstallArgs    []string
}

// List of supported package managers with their configurations
var supportedPackageManagers = []PackageManager{
	{
		Name:           "npm",
		LockFile:       "package-lock.json",
		InstallCommand: "npm",
		InstallArgs:    []string{"ci"},
	},
	{
		Name:           "yarn",
		LockFile:       "yarn.lock",
		InstallCommand: "yarn",
		InstallArgs:    []string{"install", "--frozen-lockfile"},
	},
	{
		Name:           "pnpm",
		LockFile:       "pnpm-lock.yaml",
		InstallCommand: pnpmPath,
		InstallArgs:    []string{"install", "--frozen-lockfile"},
	},
}

// IsPnpmInstalled checks if pnpm is available in the local project
func (pm *PackageManager) IsPnpmInstalled(execRunner ExecRunner) bool {
	err := execRunner.RunExecutable(pnpmPath, "--version")
	return err == nil
}

// InstallPnpm handles the special installation process for pnpm if not already installed
func (pm *PackageManager) InstallPnpm(execRunner ExecRunner) error {
	if pm.IsPnpmInstalled(execRunner) {
		log.Entry().Info("pnpm is already installed locally, skipping installation")
		return nil
	}

	if err := execRunner.RunExecutable("npm", "install", "pnpm", "--prefix", tmpInstallFolder); err != nil {
		return err
	}
	return nil
}

// detectPackageManager determines which package manager to use based on lock files
func (exec *Execute) detectPackageManager() (*PackageManager, error) {
	for _, pm := range supportedPackageManagers {
		exists, err := exec.Utils.FileExists(pm.LockFile)
		if err != nil {
			return nil, fmt.Errorf("failed to check for %s: %w", pm.LockFile, err)
		}
		if exists {
			if pm.Name == "pnpm" {
				if err := pm.InstallPnpm(exec.Utils.GetExecRunner()); err != nil {
					return nil, fmt.Errorf("failed to install pnpm: %w", err)
				}
			}
			return &pm, nil
		}
	}

	// No lock file found - log warning and default to npm with regular install
	log.Entry().Warn("No package lock file found. " +
		"It is recommended to create a `package-lock.json` file by running `npm Install` locally." +
		" Add this file to your version control. " +
		"By doing so, the builds of your application become more reliable.")

	return &PackageManager{
		Name:           "npm",
		LockFile:       "",
		InstallCommand: "npm",
		InstallArgs:    []string{"install"},
	}, nil
}
