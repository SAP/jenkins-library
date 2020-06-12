package npm

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type npmMockUtilsBundle struct {
	fileUtils  map[string]string
	execRunner mock.ExecMockRunner
}

func (u *npmMockUtilsBundle) fileExists(path string) (bool, error) {
	_, exists := u.fileUtils[path]
	return exists, nil
}

func (u *npmMockUtilsBundle) fileRead(path string) ([]byte, error) {
	return []byte(u.fileUtils[path]), nil
}

// duplicated from nexusUpload_test.go for now, refactor later?
func (u *npmMockUtilsBundle) glob(pattern string) ([]string, error) {
	var matches []string
	for path := range u.fileUtils {
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in m.fileUtils is not deterministic, this would result in flaky tests.
	sort.Sort(byLen(matches))
	return matches, nil
}

func (u *npmMockUtilsBundle) getwd() (dir string, err error) {
	return "/project", nil
}

func (u *npmMockUtilsBundle) chdir(dir string) error {
	return nil
}

func (u *npmMockUtilsBundle) getExecRunner() execRunner {
	return &u.execRunner
}

type byLen []string

func (a byLen) Len() int {
	return len(a)
}

func (a byLen) Less(i, j int) bool {
	return len(a[i]) < len(a[j])
}

func (a byLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func TestNpm(t *testing.T) {
	t.Run("find package.json files with one package.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{}"

		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageJSONFiles := exec.FindPackageJSONFiles()

		assert.Equal(t, []string{"package.json"}, packageJSONFiles)
	})

	t.Run("find package.json files with two package.json and filtered package.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{}"
		utilsMock.fileUtils["src/package.json"] = "{}"
		utilsMock.fileUtils["node_modules/package.json"] = "{}" // is filtered out
		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageJSONFiles := exec.FindPackageJSONFiles()

		assert.Equal(t, []string{"package.json", "src/package.json"}, packageJSONFiles)
	})

	t.Run("find package.json files with script", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["src/package.json"] = "{ \"name\": \"test\" }"
		utilsMock.fileUtils["test/package.json"] = "{ \"scripts\": { \"test\": \"exit 0\" } }"

		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageJSONFilesWithScript, err := exec.FindPackageJSONFilesWithScript([]string{"package.json", "src/package.json", "test/package.json"}, "ci-lint")
		assert.NoError(t, err)
		assert.Equal(t, []string{"package.json"}, packageJSONFilesWithScript)
	})

	t.Run("install deps for package.json with package-lock.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["package-lock.json"] = "{}"

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.install("package.json")
		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{"npm", []string{"ci"}}, utilsMock.execRunner.Calls[2])
	})

	t.Run("install deps for package.json without package-lock.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.install("package.json")
		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{"npm", []string{"install"}}, utilsMock.execRunner.Calls[2])
	})

	t.Run("install deps for package.json with yarn.lock", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["yarn.lock"] = "{}"

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.install("package.json")
		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{"yarn", []string{"install", "--frozen-lockfile"}}, utilsMock.execRunner.Calls[2])
	})

	t.Run("install all deps", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["package-lock.json"] = "{}"
		utilsMock.fileUtils["src/package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["src/package-lock.json"] = "{}"

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.InstallAllDependencies([]string{"package.json", "src/package.json"})
		assert.NoError(t, err)
		assert.Equal(t, 6, len(utilsMock.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{"npm", []string{"ci"}}, utilsMock.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{"npm", []string{"ci"}}, utilsMock.execRunner.Calls[5])
	})

	t.Run("check if yarn.lock and package-lock exist", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["yarn.lock"] = "{}"
		utilsMock.fileUtils["package-lock.json"] = "{}"

		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageLock, yarnLock, err := exec.checkIfLockFilesExist()
		assert.NoError(t, err)
		assert.True(t, packageLock)
		assert.True(t, yarnLock)
	})

	t.Run("check that yarn.lock and package-lock do not exist", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"

		options := executeOptions{}
		options.sapNpmRegistry = "foo.sap"
		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageLock, yarnLock, err := exec.checkIfLockFilesExist()
		assert.NoError(t, err)
		assert.False(t, packageLock)
		assert.False(t, yarnLock)
	})

	t.Run("check execute script", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"

		options := executeOptions{}
		options.install = false
		options.runScripts = []string{"ci-lint"}
		options.runOptions = []string{"--silent"}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.executeScript("package.json", "ci-lint")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utilsMock.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{"npm", []string{"run", "ci-lint", "--silent"}}, utilsMock.execRunner.Calls[2])
	})

	t.Run("check execute all scripts", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["src/package.json"] = "{\"scripts\": { \"ci-build\": \"exit 0\" } }"

		options := executeOptions{}
		options.install = false
		options.runScripts = []string{"ci-lint", "ci-build"}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.ExecuteAllScripts()
		assert.NoError(t, err)
		assert.Equal(t, 6, len(utilsMock.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{"npm", []string{"run", "ci-lint"}}, utilsMock.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{"npm", []string{"run", "ci-build"}}, utilsMock.execRunner.Calls[5])
	})

	t.Run("check set npm registry", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["src/package.json"] = "{\"scripts\": { \"ci-build\": \"exit 0\" } }"
		utilsMock.execRunner = mock.ExecMockRunner{StdoutReturn: map[string]string{"npm config get registry": "undefined"}}
		options := executeOptions{}
		options.defaultNpmRegistry = "https://example.org/npm"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.SetNpmRegistries()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utilsMock.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{"npm", []string{"config", "get", "registry"}}, utilsMock.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{"npm", []string{"config", "set", "registry", exec.options.defaultNpmRegistry}}, utilsMock.execRunner.Calls[1])
	})

	t.Run("check set npm registry", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.fileUtils["package.json"] = "{\"scripts\": { \"ci-lint\": \"exit 0\" } }"
		utilsMock.fileUtils["src/package.json"] = "{\"scripts\": { \"ci-build\": \"exit 0\" } }"
		utilsMock.execRunner = mock.ExecMockRunner{StdoutReturn: map[string]string{"npm config get @sap:registry": "undefined"}}
		options := executeOptions{}
		options.sapNpmRegistry = "https://example.sap/npm"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		err := exec.SetNpmRegistries()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utilsMock.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{"npm", []string{"config", "get", "@sap:registry"}}, utilsMock.execRunner.Calls[1])
		assert.Equal(t, mock.ExecCall{"npm", []string{"config", "set", "@sap:registry", exec.options.sapNpmRegistry}}, utilsMock.execRunner.Calls[2])
	})
}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{}
	utils.fileUtils = map[string]string{}
	return utils
}
