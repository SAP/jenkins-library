package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type npmExecuteScriptsMockUtilsBundle struct {
	execRunner mock.ExecMockRunner
	files      map[string][]byte
}

func (u *npmExecuteScriptsMockUtilsBundle) fileExists(path string) (bool, error) {
	_, exists := u.files[path]
	return exists, nil
}

// duplicated from nexusUpload_test.go for now, refactor later?
func (u *npmExecuteScriptsMockUtilsBundle) glob(pattern string) ([]string, error) {
	var matches []string
	for path := range u.files {
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in m.files is not deterministic, this would result in flaky tests.
	sort.Strings(matches)
	return matches, nil
}

func (u *npmExecuteScriptsMockUtilsBundle) getwd() (dir string, err error) {
	return "/project", nil
}

func (u *npmExecuteScriptsMockUtilsBundle) chdir(dir string) error {
	return nil
}

func (u *npmExecuteScriptsMockUtilsBundle) getExecRunner() execRunner {
	return &u.execRunner
}

func TestNpmExecuteScripts(t *testing.T) {

	t.Run("Call without install and run-scripts", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files["package-lock.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteScriptsOptions{}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(utils.execRunner.Calls))
	})

	t.Run("Project with package lock", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"
		utils.files["foo/bar/node_modules/package.json"] = "{\"name\": \"Test\" }" // is filtered out
		utils.files["gen/bar/package.json"] = "{\"name\": \"Test\" }"              // is filtered out
		utils.files["foo/gen/package.json"] = "{\"name\": \"Test\" }"              // is filtered out
		utils.files["package-lock.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
		assert.Equal(t, 3, len(npmExecutor.utils.execRunner.Calls))
	})

	t.Run("Project with two package json files", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"
		utils.files["foo/bar/package.json"] = "{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"
		utils.files["package-lock.json"] = "{\"name\": \"Test\" }"
		utils.files["foo/bar/package-lock.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, npmExecutor.utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[4])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[5])
		assert.Equal(t, 6, len(npmExecutor.utils.execRunner.Calls))
	})

	t.Run("Project with yarn lock", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"
		utils.files["yarn.lock"] = "{\"name\": \"Test\" }"
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"install", "--frozen-lockfile"}}, npmExecutor.utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
	})

	t.Run("Project without lock file", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"foo\": \"\" , \"bar\": \"\" } }"
		config := npmExecuteScriptsOptions{}
		config.Install = true
		config.RunScripts = []string{"foo", "bar"}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            config.Install,
			runScripts:         config.RunScripts,
			runOptions:         []string{},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteScripts(&npmExecutor, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install"}}, npmExecutor.utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, npmExecutor.utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "bar"}}, npmExecutor.utils.execRunner.Calls[2])
	})
}
