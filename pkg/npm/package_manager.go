package npm

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
	"path/filepath"
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
func (pm *PackageManager) IsPnpmInstalled(execRunner ExecRunner, pnpmVersion string) bool {
	err := execRunner.RunExecutable(pnpmPath, "--version")
	return err == nil
}

// InstallPnpm handles the special installation process for pnpm if not already installed
func (pm *PackageManager) InstallPnpm(execRunner ExecRunner, pnpmVersion string) error {
	if pm.IsPnpmInstalled(execRunner, pnpmVersion) {
		log.Entry().Info("pnpm is already installed locally, skipping installation. pnpmVersion parameter will be ignored.")
		return nil
	}
	installTarget := "pnpm"
	if pnpmVersion != "" {
		installTarget = fmt.Sprintf("pnpm@%s", pnpmVersion)
	}
	log.Entry().Debugf("Installing %s locally using npm", installTarget)

	if err := execRunner.RunExecutable("npm", "install", installTarget, "--prefix", tmpInstallFolder); err != nil {
		return fmt.Errorf("failed to install %s: %w", installTarget, err)
	}
	//Add pnpm installed path to PATH
	binPath := filepath.Dir(pnpmPath)
	currentPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s%c%s", binPath, os.PathListSeparator, currentPath)
	if err := os.Setenv("PATH", newPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}
	log.Entry().Debugf("Updated PATH to include %s", binPath)

	return nil
}

// detectPackageManager determines which package manager to use based on lock files
func (exec *Execute) detectPackageManager(pnpmVersion string) (*PackageManager, error) {
	for _, pm := range supportedPackageManagers {
		exists, err := exec.Utils.FileExists(pm.LockFile)
		if err != nil {
			return nil, fmt.Errorf("failed to check for %s: %w", pm.LockFile, err)
		}
		if exists {
			if pm.Name == "pnpm" {
				if err := pm.InstallPnpm(exec.Utils.GetExecRunner(), pnpmVersion); err != nil {
					return nil, fmt.Errorf("failed to install pnpm: %w", err)
				}
			}
			return &pm, nil
		}
	}

	// No lock file found - log warning and default to npm with regular install
	log.Entry().Debugf("No lock file found , default to npm with regular install")

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
