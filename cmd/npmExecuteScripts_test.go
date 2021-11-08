package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
)

// NpmMockUtilsBundle for mocking
type NpmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

// GetExecRunner return the execRunner mock
func (u *NpmMockUtilsBundle) GetExecRunner() npm.ExecRunner {
	return u.execRunner
}

// newNpmMockUtilsBundle creates an instance of NpmMockUtilsBundle
func newNpmMockUtilsBundle() NpmMockUtilsBundle {
	utils := NpmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestNpmExecuteScripts(t *testing.T) {

	cpe := npmExecuteScriptsCommonPipelineEnvironment{}

	t.Run("Call with packagesList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorList: []string{"src/package.json"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts, PackagesList: config.BuildDescriptorList}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with excludeList", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorExcludeList: []string{"**/path/**"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts, ExcludeList: config.BuildDescriptorExcludeList}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with scriptOptions", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, ScriptOptions: []string{"--run"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts, ScriptOptions: config.ScriptOptions}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call without install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: config.Install, RunScripts: config.RunScripts, VirtualFrameBuffer: config.VirtualFrameBuffer}}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("Call with createBOM", func(t *testing.T) {
		config := npmExecuteScriptsOptions{CreateBOM: true, RunScripts: []string{"ci-build", "ci-test"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}
		err := runNpmExecuteScripts(&npmExecutor, &config, &cpe)

		assert.NoError(t, err)
	})
}
