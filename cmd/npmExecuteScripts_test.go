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
	sort.Sort(byLen(matches))
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
		utils := newNpmExecuteScriptsMockUtilsBundle()
		utils.files["package.json"] = []byte(`abc`)
		utils.files["package-lock.json"] = []byte(`abc`)
		options := npmExecuteScriptsOptions{}

		err := runNpmExecuteScripts(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(utils.execRunner.Calls))
	})

	t.Run("Project with package lock", func(t *testing.T) {
		utils := newNpmExecuteScriptsMockUtilsBundle()
		utils.files["package.json"] = []byte(`abc`)
		utils.files["foo/bar/node_modules/package.json"] = []byte(`abc`) // is filtered out
		utils.files["gen/bar/package.json"] = []byte(`abc`)              // is filtered out
		utils.files["foo/gen/package.json"] = []byte(`abc`)              // is filtered out
		utils.files["package-lock.json"] = []byte(`abc`)
		options := npmExecuteScriptsOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteScripts(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[4])
		assert.Equal(t, 5, len(utils.execRunner.Calls))
	})

	t.Run("Project with two package json files", func(t *testing.T) {
		utils := newNpmExecuteScriptsMockUtilsBundle()
		utils.files["package.json"] = []byte(`abc`)
		utils.files["foo/bar/package.json"] = []byte(`abc`)
		utils.files["package-lock.json"] = []byte(`abc`)
		options := npmExecuteScriptsOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNpmExecuteScripts(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[4])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[7])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[8])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[9])
		assert.Equal(t, 10, len(utils.execRunner.Calls))
	})

	t.Run("Project with yarn lock", func(t *testing.T) {
		utils := newNpmExecuteScriptsMockUtilsBundle()
		utils.files["package.json"] = []byte(`abc`)
		utils.files["yarn.lock"] = []byte(`abc`)
		options := npmExecuteScriptsOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNpmExecuteScripts(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"install", "--frozen-lockfile"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[4])
	})

	t.Run("Project without lock file", func(t *testing.T) {
		utils := newNpmExecuteScriptsMockUtilsBundle()
		utils.files["package.json"] = []byte(`abc`)
		options := npmExecuteScriptsOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNpmExecuteScripts(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[4])
	})
}

func newNpmExecuteScriptsMockUtilsBundle() npmExecuteScriptsMockUtilsBundle {
	utils := npmExecuteScriptsMockUtilsBundle{}
	utils.files = map[string][]byte{}
	return utils
}
