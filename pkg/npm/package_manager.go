package npm

import (
	"fmt"
	"os"
	"path/filepath"

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
func (pm *PackageManager) InstallPnpm(execRunner ExecRunner, pnpmVersion string, rootDir string) error {
	pnpmPackage := "pnpm"
	if pnpmVersion != "" {
		pnpmPackage = fmt.Sprintf("pnpm@%s", pnpmVersion)
	}

	// Calculate the prefix path relative to root directory
	prefixPath := filepath.Join(rootDir, tmpInstallFolder)

	if err := execRunner.RunExecutable("npm", "install", pnpmPackage, "--prefix", prefixPath); err != nil {
		return fmt.Errorf("failed to install pnpm locally: %w", err)
	}

	// Add the local pnpm bin directory to PATH so it's globally available
	pnpmBinDir := filepath.Join(prefixPath, "node_modules", ".bin")
	currentPath := os.Getenv("PATH")
	newPath := pnpmBinDir + string(os.PathListSeparator) + currentPath
	os.Setenv("PATH", newPath)

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
				if err := exec.setupPnpm(&pm); err != nil {
					return nil, fmt.Errorf("failed to set up pnpm: %w", err)
				}
				// Use cached pnpm command
				pm.InstallCommand = exec.pnpmSetup.command
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

// setupPnpm handles the pnpm installation and setup process
func (exec *Execute) setupPnpm(pm *PackageManager) error {
	// Check pnpm installation only once per Execute instance
	if !exec.pnpmSetup.checked {
		exec.pnpmSetup.checked = true
		execRunner := exec.Utils.GetExecRunner()

		// Set up local pnpm path if not already set
		if exec.pnpmSetup.command == "" {
			absolutePnpmPath := filepath.Join(exec.pnpmSetup.rootDir, tmpInstallFolder, "node_modules", ".bin", "pnpm")
			exec.pnpmSetup.command = absolutePnpmPath
		}

		// If a specific pnpm version is requested, always install and use it locally
		if exec.Options.PnpmVersion != "" {
			// Install the specified pnpm version locally
			if err := pm.InstallPnpm(execRunner, exec.Options.PnpmVersion, exec.pnpmSetup.rootDir); err != nil {
				return fmt.Errorf("failed to install specified pnpm version: %w", err)
			}
			exec.pnpmSetup.installed = true
			log.Entry().Infof("Using locally installed pnpm version %s", exec.Options.PnpmVersion)
		} else {
			// Check if pnpm is globally installed first
			if err := execRunner.RunExecutable("pnpm", "--version"); err == nil {
				// Use global pnpm
				exec.pnpmSetup.installed = true
				exec.pnpmSetup.command = "pnpm"
				log.Entry().Info("Using globally installed pnpm")
			} else {
				// Check if pnpm is locally installed at root
				if err := execRunner.RunExecutable(exec.pnpmSetup.command, "--version"); err == nil {
					// Use local pnpm
					exec.pnpmSetup.installed = true
					log.Entry().Info("Using locally installed pnpm")
				} else {
					// Install pnpm locally with configured version (only once, in root directory)
					if err := pm.InstallPnpm(execRunner, "", exec.pnpmSetup.rootDir); err != nil {
						return fmt.Errorf("failed to install specified pnpm version: %w", err)
					}
					// Use local pnpm after installation
					exec.pnpmSetup.installed = true
					log.Entry().Info("Using locally installed pnpm")
				}
			}
		}
	}
	return nil
}
