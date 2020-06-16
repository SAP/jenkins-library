package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestNpmExecuteScripts(t *testing.T) {

	t.Run("Call without install and run-scripts", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile("package-lock.json", []byte("{\"name\": \"Test\" }"))
		config := npmExecuteScriptsOptions{}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			assert.Equal(t, 0, len(utils.execRunner.Calls))
		}
	})

	t.Run("Project with package lock", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"))
		utils.AddFile(filepath.Join("foo", "bar", "node_modules", "package.json"), []byte("{\"name\": \"Test\" }")) // is filtered out
		utils.AddFile(filepath.Join("gen", "bar", "package.json"), []byte("{\"name\": \"Test\" }"))                 // is filtered out
		utils.AddFile(filepath.Join("foo", "gen", "package.json"), []byte("{\"name\": \"Test\" }"))                 // is filtered out
		utils.AddFile("package-lock.json", []byte("{\"name\": \"Test\" }"))
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(npmExecutor.utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Project with two package json files", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"))
		utils.AddFile(filepath.Join("foo", "bar", "package.json"), []byte("{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"))
		utils.AddFile("package-lock.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(filepath.Join("foo", "bar", "package-lock.json"), []byte("{\"name\": \"Test\" }"))
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(npmExecutor.utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[3])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[4])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[5])
			}
		}
	})

	t.Run("Project with yarn lock", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"))
		utils.AddFile("yarn.lock", []byte("{\"name\": \"Test\" }"))
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(npmExecutor.utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"install", "--frozen-lockfile"}}, npmExecutor.utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Project without lock file", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"))
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(npmExecutor.utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install"}}, npmExecutor.utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
			}
		}
	})
}
