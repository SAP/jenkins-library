//go:build !release
// +build !release

package npm

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/SAP/jenkins-library/pkg/mock"
)

// NpmMockUtilsBundle for mocking
type NpmMockUtilsBundle struct {
	*mock.FilesMock
	ExecRunner *mock.ExecMockRunner
}

// GetExecRunner return the execRunner mock
func (u *NpmMockUtilsBundle) GetExecRunner() ExecRunner {
	return u.ExecRunner
}

// NewNpmMockUtilsBundle creates an instance of NpmMockUtilsBundle
func NewNpmMockUtilsBundle() NpmMockUtilsBundle {
	utils := NpmMockUtilsBundle{FilesMock: &mock.FilesMock{}, ExecRunner: &mock.ExecMockRunner{}}
	return utils
}

// NpmConfig holds the config parameters needed for checking if the function is called with correct parameters
type NpmConfig struct {
	Install            bool
	RunScripts         []string
	RunOptions         []string
	ScriptOptions      []string
	VirtualFrameBuffer bool
	ExcludeList        []string
	PackagesList       []string
}

// NpmExecutorMock mocking struct
type NpmExecutorMock struct {
	Utils  NpmMockUtilsBundle
	Config NpmConfig
}

// FindPackageJSONFiles mock implementation
func (n *NpmExecutorMock) FindPackageJSONFiles() []string {
	packages, _ := n.Utils.Glob("**/package.json")
	return packages
}

// FindPackageJSONFiles mock implementation
func (n *NpmExecutorMock) FindPackageJSONFilesWithExcludes(excludeList []string) ([]string, error) {
	packages, _ := n.Utils.Glob("**/package.json")
	return packages, nil
}

// FindPackageJSONFilesWithScript mock implementation
func (n *NpmExecutorMock) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	return packageJSONFiles, nil
}

// RunScriptsInAllPackages mock implementation
func (n *NpmExecutorMock) RunScriptsInAllPackages(runScripts []string, runOptions []string, scriptOptions []string, virtualFrameBuffer bool, excludeList []string, packagesList []string) error {
	if len(runScripts) != len(n.Config.RunScripts) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.runScripts")
	}
	for i, script := range runScripts {
		if script != n.Config.RunScripts[i] {
			return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.runScripts")
		}
	}

	if len(scriptOptions) != len(n.Config.ScriptOptions) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of scriptOptions than config.scriptOptions")
	}

	if len(runOptions) != len(n.Config.RunOptions) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runOptions than config.runOptions")
	}

	if virtualFrameBuffer != n.Config.VirtualFrameBuffer {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of virtualFrameBuffer than config.virtualFrameBuffer")
	}

	if len(excludeList) != len(n.Config.ExcludeList) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of excludeList than config.excludeList")
	}

	if len(packagesList) != len(n.Config.PackagesList) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of packagesList than config.packagesList")
	}

	return nil
}

// InstallAllDependencies mock implementation
func (n *NpmExecutorMock) InstallAllDependencies(packageJSONFiles []string, pnpmVersion string) error {
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
		return fmt.Errorf("InstallAllDependencies was called but config.install was false")
	}
	return nil
}

// SetNpmRegistries mock implementation
func (n *NpmExecutorMock) SetNpmRegistries() error {
	return nil
}

// CreateBOM mock implementation
func (n *NpmExecutorMock) CreateBOM(packageJSONFiles []string) error {
	return nil
}

// CreateBOM mock implementation
func (n *NpmExecutorMock) PublishAllPackages(packageJSONFiles []string, registry, username, password string, packBeforePublish bool, buildCoordinates *[]versioning.Coordinates) error {
	return nil
}
