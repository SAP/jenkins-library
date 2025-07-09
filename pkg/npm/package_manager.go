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

// InstallPnpm handles the special installation process for pnpm locally
func (pm *PackageManager) InstallPnpm(execRunner ExecRunner, pnpmVersion string) error {
	pnpmPackage := "pnpm"
	if pnpmVersion != "" && pnpmVersion != "latest" {
		pnpmPackage = fmt.Sprintf("pnpm@%s", pnpmVersion)
	}

	if err := execRunner.RunExecutable("npm", "install", pnpmPackage, "--prefix", tmpInstallFolder); err != nil {
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
				execRunner := exec.Utils.GetExecRunner()

				// Check if pnpm is globally installed first
				if err := execRunner.RunExecutable("pnpm", "--version"); err == nil {
					// Use global pnpm
					pm.InstallCommand = "pnpm"
					log.Entry().Info("Using globally installed pnpm")
				} else {
					// Check if pnpm is locally installed
					if err := execRunner.RunExecutable(pnpmPath, "--version"); err == nil {
						// Use local pnpm
						pm.InstallCommand = pnpmPath
						log.Entry().Info("Using locally installed pnpm")
					} else {
						// Install pnpm locally with configured version
						if err := pm.InstallPnpm(execRunner, exec.Options.PnpmVersion); err != nil {
							return nil, fmt.Errorf("failed to install pnpm: %w", err)
						}
						// Use local pnpm after installation
						pm.InstallCommand = pnpmPath
						log.Entry().Info("Using locally installed pnpm")
					}
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
