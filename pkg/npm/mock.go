// +build !release

package npm

import (
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
	FoundPackageFiles  []string
}

// NpmExecutorMock mocking struct
type NpmExecutorMock struct {
	Utils    NpmMockUtilsBundle
	Received NpmConfig
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
	n.Received.RunScripts = runScripts
	n.Received.ScriptOptions = scriptOptions
	n.Received.RunOptions = runOptions
	n.Received.VirtualFrameBuffer = virtualFrameBuffer
	n.Received.PackagesList = packagesList
	n.Received.ExcludeList = excludeList
	return nil
}

// InstallAllDependencies mock implementation
func (n *NpmExecutorMock) InstallAllDependencies(packageJSONFiles []string) error {
	n.Received.FoundPackageFiles = packageJSONFiles
	n.Received.Install = true
	return nil
}

// SetNpmRegistries mock implementation
func (n *NpmExecutorMock) SetNpmRegistries() error {
	return nil
}
