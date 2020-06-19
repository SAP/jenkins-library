package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
	"testing"
)

type npmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

func (u *npmMockUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.execRunner
}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

type npmExecutorMock struct {
	utils  npmMockUtilsBundle
	config npmExecuteScriptsOptions
}

func (n *npmExecutorMock) FindPackageJSONFiles() []string {
	packages, _ := n.utils.Glob("**/package.json")
	return packages
}

func (n *npmExecutorMock) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	return packageJSONFiles, nil
}

func (n *npmExecutorMock) RunScriptsInAllPackages(runScripts []string, runOptions []string, virtualFrameBuffer bool) error {
	if len(runScripts) != len(n.config.RunScripts) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.RunScripts")
	}
	for i, script := range runScripts {
		if script != n.config.RunScripts[i] {
			return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.RunScripts")
		}
	}

	if len(runOptions) != 0 {
		return fmt.Errorf("RunScriptsInAllPackages was unexpectedly called with a list of runOptions")
	}

	if virtualFrameBuffer != n.config.VirtualFrameBuffer {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of virtualFrameBuffer than config.virtualFrameBuffer")
	}

	return nil
}

func (n *npmExecutorMock) InstallAllDependencies(packageJSONFiles []string) error {
	allPackages := n.FindPackageJSONFiles()
	if len(packageJSONFiles) != len(allPackages) {
		return fmt.Errorf("packageJSONFiles != n.FindPackageJSONFiles()")
	}
	for i, packageJSON := range packageJSONFiles {
		if packageJSON != allPackages[i] {
			return fmt.Errorf("InstallAllDependencies was called with a different list of package.json files than result of n.FindPackageJSONFiles()")
		}
	}

	if !n.config.Install {
		return fmt.Errorf("InstallAllDependencies was called but config.Install was false")
	}
	return nil
}

func (n *npmExecutorMock) SetNpmRegistries() error {
	return nil
}

func TestNpmExecuteScripts(t *testing.T) {
	t.Run("Call with install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: config}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call without install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: config}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: config}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{SapNpmRegistry: config.SapNpmRegistry, DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "@sap:registry"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.execRunner.Calls[5])
			}
		}
	})
}
