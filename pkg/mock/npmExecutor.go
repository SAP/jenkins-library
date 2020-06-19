package mock

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/npm"
)

type NpmUtilsBundle struct {
	*FilesMock
	ExecRunner *ExecMockRunner
}

func (u *NpmUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.ExecRunner
}

func NewNpmUtilsBundle() NpmUtilsBundle {
	utils := NpmUtilsBundle{FilesMock: &FilesMock{}, ExecRunner: &ExecMockRunner{}}
	return utils
}

type NpmConfig struct {
	Install            bool
	RunScripts         []string
	RunOptions         []string
	VirtualFrameBuffer bool
}

type NpmExecutor struct {
	Utils  NpmUtilsBundle
	Config NpmConfig
}

func (n *NpmExecutor) FindPackageJSONFiles() []string {
	packages, _ := n.Utils.Glob("**/package.json")
	return packages
}

func (n *NpmExecutor) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	return packageJSONFiles, nil
}

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

func (n *NpmExecutor) SetNpmRegistries() error {
	return nil
}
