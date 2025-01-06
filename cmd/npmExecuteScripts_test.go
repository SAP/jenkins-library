//go:build unit
// +build unit

package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
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
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorList: []string{"package.json", "src/package.json"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts, PackagesList: cfg.BuildDescriptorList}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with excludeList", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, BuildDescriptorExcludeList: []string{"**/path/**"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts, ExcludeList: cfg.BuildDescriptorExcludeList}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with scriptOptions", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, ScriptOptions: []string{"--run"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts, ScriptOptions: cfg.ScriptOptions}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with install", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call without install", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := npm.NewNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.NpmExecutorMock{Utils: utils, Config: npm.NpmConfig{Install: cfg.Install, RunScripts: cfg.RunScripts, VirtualFrameBuffer: cfg.VirtualFrameBuffer}}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: cfg.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("Call with createBOM", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{CreateBOM: true, RunScripts: []string{"ci-build", "ci-test"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: cfg.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.Execute{Utils: &utils, Options: options}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)

		assert.NoError(t, err)
	})

	t.Run("Call with production", func(t *testing.T) {
		cfg := npmExecuteScriptsOptions{Production: true, RunScripts: []string{"ci-build", "ci-test"}}

		options := npm.ExecutorOptions{DefaultNpmRegistry: cfg.DefaultNpmRegistry}

		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		SetConfigOptions(ConfigCommandOptions{
			OpenFile: config.OpenPiperFile,
		})

		npmExecutor := npm.Execute{Utils: &utils, Options: options}
		err := runNpmExecuteScripts(&npmExecutor, &cfg, &cpe)
		assert.NoError(t, err)

		v := os.Getenv("NODE_ENV")
		assert.Equal(t, "production", v)
	})

}
