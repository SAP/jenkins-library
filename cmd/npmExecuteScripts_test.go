package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNpmExecuteScripts(t *testing.T) {
	t.Run("Call with install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := mock.NewNpmUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := mock.NpmExecutor{Utils: utils, Config: mock.NpmConfig{Install: config.Install, RunScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call without install", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}}
		utils := mock.NewNpmUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := mock.NpmExecutor{Utils: utils, Config: mock.NpmConfig{Install: config.Install, RunScripts: config.RunScripts}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with virtualFrameBuffer", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build", "ci-test"}, VirtualFrameBuffer: true}
		utils := mock.NewNpmUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("src/package.json", []byte("{\"name\": \"Test\" }"))

		npmExecutor := mock.NpmExecutor{Utils: utils, Config: mock.NpmConfig{Install: config.Install, RunScripts: config.RunScripts, VirtualFrameBuffer: config.VirtualFrameBuffer}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
	})

	t.Run("Test integration with npm pkg", func(t *testing.T) {
		config := npmExecuteScriptsOptions{Install: true, RunScripts: []string{"ci-build"}}

		options := npm.ExecutorOptions{SapNpmRegistry: config.SapNpmRegistry, DefaultNpmRegistry: config.DefaultNpmRegistry}

		utils := mock.NewNpmUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-build\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte(""))

		npmExecutor := npm.Execute{Utils: &utils, Options: options}

		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(utils.ExecRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utils.ExecRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "@sap:registry"}}, utils.ExecRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.ExecRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.ExecRunner.Calls[5])
			}
		}
	})
}
