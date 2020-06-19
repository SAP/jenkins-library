package mock

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/npm"
)

// NpmUtilsBundle for mocking
type NpmUtilsBundle struct {
	*FilesMock
	ExecRunner *ExecMockRunner
}

// GetExecRunner return the execRunner mock
func (u *NpmUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.ExecRunner
}

// NewNpmUtilsBundle creates an instance of NpmUtilsBundle
func NewNpmUtilsBundle() NpmUtilsBundle {
	utils := NpmUtilsBundle{FilesMock: &FilesMock{}, ExecRunner: &ExecMockRunner{}}
	return utils
}

// NpmConfig holds the config parameters needed for checking if the function is called with correct parameters
type NpmConfig struct {
	Install            bool
	RunScripts         []string
	RunOptions         []string
	VirtualFrameBuffer bool
}

// NpmExecutor mocking struct
type NpmExecutor struct {
	Utils  NpmUtilsBundle
	Config NpmConfig
}

// FindPackageJSONFiles mock implementation
func (n *NpmExecutor) FindPackageJSONFiles() []string {
	packages, _ := n.Utils.Glob("**/package.json")
	return packages
}

// FindPackageJSONFilesWithScript mock implementation
func (n *NpmExecutor) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	return packageJSONFiles, nil
}

// RunScriptsInAllPackages mock implementation
func (n *NpmExecutor) RunScriptsInAllPackages(runScripts []string, runOptions []string, virtualFrameBuffer bool) error {
	if len(runScripts) != len(n.Config.RunScripts) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.RunScripts")
	}
	for i, script := range runScripts {
		if script != n.Config.RunScripts[i] {
			return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.RunScripts")
		}
	}

	if len(runOptions) != 0 {
		return fmt.Errorf("RunScriptsInAllPackages was unexpectedly called with a list of runOptions")
	}

	if virtualFrameBuffer != n.Config.VirtualFrameBuffer {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of virtualFrameBuffer than config.virtualFrameBuffer")
	}

	return nil
}

// InstallAllDependencies mock implementation
func (n *NpmExecutor) InstallAllDependencies(packageJSONFiles []string) error {
	allPackages := n.FindPackageJSONFiles()
	if len(packageJSONFiles) != len(allPackages) {
		return fmt.Errorf("packageJSONFiles != n.FindPackageJSONFiles()")
	}
	for i, packageJSON := range packageJSONFiles {
		if packageJSON != allPackages[i] {
			return fmt.Errorf("InstallAllDependencies was called with a different list of package.json files than result of n.FindPackageJSONFiles()")
		}
	}

	if !n.Config.Install {
		return fmt.Errorf("InstallAllDependencies was called but config.Install was false")
	}
	return nil
}

// SetNpmRegistries mock implementation
func (n *NpmExecutor) SetNpmRegistries() error {
	return nil
}
