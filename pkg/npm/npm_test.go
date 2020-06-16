package npm

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

type npmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner mock.ExecMockRunner
}

func (u *npmMockUtilsBundle) getExecRunner() execRunner {
	return &u.execRunner
}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}}
	return utils
}

func TestNpm(t *testing.T) {
	t.Run("find package.json files with one package.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"name\": \"Test\" }"))

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
		utilsMock.AddFile("package.json", []byte("{}"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{}"))
		utilsMock.AddFile(filepath.Join("node_modules", "package.json"), []byte("{}")) // is filtered out
		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageJSONFiles := exec.FindPackageJSONFiles()

		assert.Equal(t, []string{"package.json", filepath.Join("src", "package.json")}, packageJSONFiles)
	})

	t.Run("find package.json files with script", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{ \"name\": \"test\" }"))
		utilsMock.AddFile(filepath.Join("test", "package.json"), []byte("{ \"scripts\": { \"test\": \"exit 0\" } }"))

		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}

		packageJSONFilesWithScript, err := exec.FindPackageJSONFilesWithScript([]string{"package.json", filepath.Join("src", "package.json"), filepath.Join("test", "package.json")}, "ci-lint")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"package.json"}, packageJSONFilesWithScript)
		}
	})

	t.Run("Install deps for package.json with package-lock.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile("package-lock.json", []byte("{}"))

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utilsMock.execRunner.Calls[2])
			}
		}
	})

	t.Run("Install deps for package.json without package-lock.json", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"Install"}}, utilsMock.execRunner.Calls[2])
			}
		}
	})

	t.Run("Install deps for package.json with yarn.lock", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile("yarn.lock", []byte("{}"))

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"Install", "--frozen-lockfile"}}, utilsMock.execRunner.Calls[2])
			}
		}
	})

	t.Run("Install all deps", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile("package-lock.json", []byte("{}"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))

		options := executeOptions{}
		options.defaultNpmRegistry = "foo.bar"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.InstallAllDependencies([]string{"package.json", filepath.Join("src", "package.json")})

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utilsMock.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utilsMock.execRunner.Calls[5])
			}
		}
	})

	t.Run("check if yarn.lock and package-lock exist", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile("yarn.lock", []byte("{}"))
		utilsMock.AddFile("package-lock.json", []byte("{}"))

		options := executeOptions{}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		packageLock, yarnLock, err := exec.checkIfLockFilesExist()

		if assert.NoError(t, err) {
			assert.True(t, packageLock)
			assert.True(t, yarnLock)
		}
	})

	t.Run("check that yarn.lock and package-lock do not exist", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := executeOptions{}
		options.sapNpmRegistry = "foo.sap"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		packageLock, yarnLock, err := exec.checkIfLockFilesExist()

		if assert.NoError(t, err) {
			assert.False(t, packageLock)
			assert.False(t, yarnLock)
		}
	})

	t.Run("check execute script", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := executeOptions{}
		options.install = false
		options.runScripts = []string{"ci-lint"}
		options.runOptions = []string{"--silent"}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.executeScript("package.json", "ci-lint")

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utilsMock.execRunner.Calls[2])
			}
		}
	})

	t.Run("check execute all scripts", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))

		options := executeOptions{}
		options.install = false
		options.runScripts = []string{"ci-lint", "ci-build"}

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.ExecuteAllScripts()

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint"}}, utilsMock.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utilsMock.execRunner.Calls[5])
			}
		}
	})

	t.Run("check set npm registry", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))
		utilsMock.execRunner = mock.ExecMockRunner{StdoutReturn: map[string]string{"npm config get registry": "undefined"}}
		options := executeOptions{}
		options.defaultNpmRegistry = "https://example.org/npm"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.SetNpmRegistries()

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry"}}, utilsMock.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "set", "registry", exec.options.defaultNpmRegistry}}, utilsMock.execRunner.Calls[1])
			}
		}
	})

	t.Run("check set npm registry", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utilsMock.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))
		utilsMock.execRunner = mock.ExecMockRunner{StdoutReturn: map[string]string{"npm config get @sap:registry": "undefined"}}
		options := executeOptions{}
		options.sapNpmRegistry = "https://example.sap/npm"

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.SetNpmRegistries()

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utilsMock.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "@sap:registry"}}, utilsMock.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "set", "@sap:registry", exec.options.sapNpmRegistry}}, utilsMock.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call run-scripts with virtual frame buffer", func(t *testing.T) {
		utilsMock := newNpmMockUtilsBundle()
		utilsMock.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" } }"))

		options := executeOptions{}
		options.install = false
		options.runScripts = []string{"foo"}
		options.virtualFrameBuffer = true

		exec := &execute{
			utils:   &utilsMock,
			options: options,
		}
		err := exec.ExecuteAllScripts()

		assert.Contains(t, utilsMock.execRunner.Env, "DISPLAY=:99")
		assert.NoError(t, err)
		if assert.Len(t, utilsMock.execRunner.Calls, 4) {
			xvfbCall := utilsMock.execRunner.Calls[0]
			assert.Equal(t, "Xvfb", xvfbCall.Exec)
			assert.Equal(t, []string{"-ac", ":99", "-screen", "0", "1280x1024x16"}, xvfbCall.Params)
			assert.True(t, xvfbCall.Async)
			assert.True(t, xvfbCall.Execution.Killed)

			assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, utilsMock.execRunner.Calls[3])
		}
	})
}
