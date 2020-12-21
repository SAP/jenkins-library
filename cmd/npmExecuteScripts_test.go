package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
	"testing"
)

// npmMockUtilsBundle for mocking
type npmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

// GetExecRunner return the execRunner mock
func (u *npmMockUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.execRunner
}

// newNpmMockUtilsBundle creates an instance of npmMockUtilsBundle
func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

// npmConfig holds the config parameters needed for checking if the function is called with correct parameters
type npmConfig struct {
	install            bool
	runScripts         []string
	runOptions         []string
	scriptOptions      []string
	virtualFrameBuffer bool
	excludeList        []string
	packagesList       []string
}

// npmExecutorMock mocking struct
type npmExecutorMock struct {

	utils  npmMockUtilsBundle
	config npmConfig
	// config above should disappear with other PR
	conf   []string
	calls [][]string
}

func (exec *npmExecutorMock) SetConfig(key, value string) error {
	// finally we mock here how the real impl behaves. Actually we shoulnd't to this here
	// this is just for outlining. Later on we should simply collect the values we receive.
	if len(value) > 0 {
		exec.conf = append(exec.conf, fmt.Sprintf("--%s=%s", key, value))
	} else {
		exec.conf = append(exec.conf, fmt.Sprintf("--%s", key))
	}
	return nil
}

func (exec *npmExecutorMock) Execute(args []string) error {
	// finally we mock here how the real impl behaves. Actually we shoulnd't to this here
	// this is just for outlining. Later on we should simply collect the values we receive.
	call := []string{"npm"}
	call = append(call, exec.conf...)
	call = append(call, args...)
	exec.calls = append(exec.calls, call)
	return nil
}

// FindPackageJSONFiles mock implementation
func (n *npmExecutorMock) FindPackageJSONFiles() []string {
	packages, _ := n.utils.Glob("**/package.json")
	return packages
}

// FindPackageJSONFiles mock implementation
func (n *npmExecutorMock) FindPackageJSONFilesWithExcludes(excludeList []string) ([]string, error) {
	packages, _ := n.utils.Glob("**/package.json")
	return packages, nil
}

// FindPackageJSONFilesWithScript mock implementation
func (n *npmExecutorMock) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	return packageJSONFiles, nil
}

// RunScriptsInAllPackages mock implementation
func (n *npmExecutorMock) RunScriptsInAllPackages(runScripts []string, runOptions []string, scriptOptions []string, virtualFrameBuffer bool, excludeList []string, packagesList []string) error {
	if len(runScripts) != len(n.config.runScripts) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.runScripts")
	}
	for i, script := range runScripts {
		if script != n.config.runScripts[i] {
			return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runScripts than config.runScripts")
		}
	}

	if len(scriptOptions) != len(n.config.scriptOptions) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of scriptOptions than config.scriptOptions")
	}

	if len(runOptions) != len(n.config.runOptions) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different list of runOptions than config.runOptions")
	}

	if virtualFrameBuffer != n.config.virtualFrameBuffer {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of virtualFrameBuffer than config.virtualFrameBuffer")
	}

	if len(excludeList) != len(n.config.excludeList) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of excludeList than config.excludeList")
	}

	if len(packagesList) != len(n.config.packagesList) {
		return fmt.Errorf("RunScriptsInAllPackages was called with a different value of packagesList than config.packagesList")
	}

	return nil
}

// InstallAllDependencies mock implementation
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

	if !n.config.install {
		return fmt.Errorf("InstallAllDependencies was called but config.install was false")
	}
	return nil
}

// SetNpmRegistries mock implementation
func (n *npmExecutorMock) SetNpmRegistries() error {
	return nil
}

func TestNpmExecuteScripts(t *testing.T) {
	t.Run("Call with packagesList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorList: []string{"src/package.json"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts, packagesList: config.BuildDescriptorList}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with excludeList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorExcludeList: []string{"**/path/**"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts, excludeList: config.BuildDescriptorExcludeList}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with scriptOptions", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, ScriptOptions: []string{"--run"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts, scriptOptions: config.ScriptOptions}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call without install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{install: config.Install, runScripts: config.RunScripts, virtualFrameBuffer: config.VirtualFrameBuffer}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"--registry=https://registry.npmjs.org", "ci"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"--registry=https://registry.npmjs.org", "run", "ci-build"}}, utils.execRunner.Calls[1])
			}
		}
	})

	//
	// This test does basically not belong to here. It is only here since we can't access the mock from pkg currently.
	t.Run("Execute", func(t *testing.T) {
                //config := npmExecuteScriptsOptions{}
                utils := newNpmMockUtilsBundle()
                npmExecutor := npmExecutorMock{utils: utils, config: npmConfig{}}
		npmExecutor.SetConfig("@mycomp:registry", "http://npm.my-company.com")
		npmExecutor.SetConfig("verbose", "")
		err := npmExecutor.Execute([]string{"install", "--global"})
		if assert.NoError(t, err) {
			// The mock here mimics the behaviour of the real impl. That is of course not that cool. Finally the mock should "only"
			// remember the args and the config entries. We should finally assert only this, not the call. The (full) call needs to
			// be asserted on the level of the exec runner. This here is only for outlining ...
			assert.Equal(t, []string{"npm", "--@mycomp:registry=http://npm.my-company.com", "--verbose", "install", "--global"}, npmExecutor.calls[0])
		}
	})
}
