package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type nodeJsBuildMockUtilsBundle struct {
	execRunner mock.ExecMockRunner
	files      map[string][]byte
}

func (u *nodeJsBuildMockUtilsBundle) fileExists(path string) (bool, error) {
	_, exists := u.files[path]
	return exists, nil
}

// duplicated from nexusUpload_test.go for now, refactor later?
func (u *nodeJsBuildMockUtilsBundle) glob(pattern string) ([]string, error) {
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

func (u *nodeJsBuildMockUtilsBundle) getwd() (dir string, err error) {
	return "/project", nil
}

func (u *nodeJsBuildMockUtilsBundle) dir(fileName string) string {
	return "/project"
}

func (u *nodeJsBuildMockUtilsBundle) chdir(dir string) error {
	return nil
}

func (u *nodeJsBuildMockUtilsBundle) getExecRunner() execRunner {
	return &u.execRunner
}

func TestNodeJsBuild(t *testing.T) {
	t.Run("Call without install and run-scripts", func(t *testing.T) {
		utils := nodeJsBuildMockUtilsBundle{}
		utils.files = map[string][]byte{}
		utils.files["package.json"] = []byte(`abc`)
		utils.files["package-lock.json"] = []byte(`abc`)
		options := nodeJsBuildOptions{}

		err := runNodeJsBuild(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(utils.execRunner.Calls))
	})

	t.Run("Project with package lock", func(t *testing.T) {
		utils := nodeJsBuildMockUtilsBundle{}
		utils.files = map[string][]byte{}
		utils.files["package.json"] = []byte(`abc`)
		utils.files["foo/bar/node_modules/package.json"] = []byte(`abc`) // is filtered out
		utils.files["package-lock.json"] = []byte(`abc`)
		options := nodeJsBuildOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNodeJsBuild(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[2])
		assert.Equal(t, 3, len(utils.execRunner.Calls))
		assert.Equal(t, []string{"npm_config_@sap:registry=", "npm_config_registry=foo.bar"},  utils.execRunner.Env)
	})

	t.Run("Project with two package lock files", func(t *testing.T) {
		utils := nodeJsBuildMockUtilsBundle{}
		utils.files = map[string][]byte{}
		utils.files["package.json"] = []byte(`abc`)
		utils.files["foo/bar/package.json"] = []byte(`abc`)
		utils.files["package-lock.json"] = []byte(`abc`)
		options := nodeJsBuildOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNodeJsBuild(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[3])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[4])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[5])
		assert.Equal(t, 6, len(utils.execRunner.Calls))
	})

	t.Run("Project with yarn lock", func(t *testing.T) {
		utils := nodeJsBuildMockUtilsBundle{}
		utils.files = map[string][]byte{}
		utils.files["package.json"] = []byte(`abc`)
		utils.files["yarn.lock"] = []byte(`abc`)
		options := nodeJsBuildOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNodeJsBuild(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"install", "--frozen-lockfile"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[2])
	})

	t.Run("Project without lock file", func(t *testing.T) {
		utils := nodeJsBuildMockUtilsBundle{}
		utils.files = map[string][]byte{}
		utils.files["package.json"] = []byte(`abc`)
		options := nodeJsBuildOptions{}
		options.Install = true
		options.RunScripts = []string{"foo", "bar"}

		err := runNodeJsBuild(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "foo", "--if-present"}}, utils.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run-script", "bar", "--if-present"}}, utils.execRunner.Calls[2])
	})
}
