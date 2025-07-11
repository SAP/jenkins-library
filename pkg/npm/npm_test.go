//go:build unit
// +build unit

package npm

import (
	"fmt"
	"path/filepath"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

type npmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner     *mock.ExecMockRunner
	downloadClient *mock.HttpClientMock
}

func (u *npmMockUtilsBundle) GetExecRunner() ExecRunner {
	return u.execRunner
}

func (u *npmMockUtilsBundle) GetFileUtils() piperutils.FileUtils {
	return u.FilesMock
}

func (u *npmMockUtilsBundle) GetDownloadUtils() piperhttp.Downloader {
	return u.downloadClient
}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	filesMock := &mock.FilesMock{}
	utils := npmMockUtilsBundle{
		FilesMock:  filesMock,
		execRunner: &mock.ExecMockRunner{},
		downloadClient: &mock.HttpClientMock{
			FileUploads:   make(map[string]string),
			HTTPFileUtils: filesMock,
		},
	}
	return utils
}

func TestNpm(t *testing.T) {
	t.Run("find package.json files with one package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}

		packageJSONFiles := exec.FindPackageJSONFiles()

		assert.Equal(t, []string{"package.json"}, packageJSONFiles)
	})

	t.Run("find package.json files with two package.json and default filter", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{}"))          // should NOT be filtered out
		utils.AddFile(filepath.Join("node_modules", "package.json"), []byte("{}")) // is filtered out
		utils.AddFile(filepath.Join("gen", "package.json"), []byte("{}"))          // is filtered out

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}

		packageJSONFiles := exec.FindPackageJSONFiles()

		assert.Equal(t, []string{"package.json", filepath.Join("src", "package.json")}, packageJSONFiles)
	})

	t.Run("find package.json files with two package.json and excludes", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{}"))                  // should NOT be filtered out
		utils.AddFile(filepath.Join("notfiltered", "package.json"), []byte("{}"))          // should NOT be filtered out
		utils.AddFile(filepath.Join("Path", "To", "filter", "package.json"), []byte("{}")) // should NOT be filtered out
		utils.AddFile(filepath.Join("node_modules", "package.json"), []byte("{}"))         // is filtered out
		utils.AddFile(filepath.Join("gen", "package.json"), []byte("{}"))                  // is filtered out
		utils.AddFile(filepath.Join("filter", "package.json"), []byte("{}"))               // is filtered out
		utils.AddFile(filepath.Join("filterPath", "package.json"), []byte("{}"))           // is filtered out
		utils.AddFile(filepath.Join("filter", "Path", "To", "package.json"), []byte("{}")) // is filtered out

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}

		packageJSONFiles, err := exec.FindPackageJSONFilesWithExcludes([]string{"filter/**", "filterPath/package.json"})

		if assert.NoError(t, err) {
			assert.Equal(t, []string{filepath.Join("Path", "To", "filter", "package.json"), filepath.Join("notfiltered", "package.json"), "package.json", filepath.Join("src", "package.json")}, packageJSONFiles)
		}
	})

	t.Run("find package.json files with script", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{ \"name\": \"test\" }"))
		utils.AddFile(filepath.Join("test", "package.json"), []byte("{ \"scripts\": { \"test\": \"exit 0\" } }"))

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}

		packageJSONFilesWithScript, err := exec.FindPackageJSONFilesWithScript([]string{"package.json", filepath.Join("src", "package.json"), filepath.Join("test", "package.json")}, "ci-lint")

		if assert.NoError(t, err) {
			assert.Equal(t, []string{"package.json"}, packageJSONFilesWithScript)
		}
	})

	t.Run("Install deps for package.json with package-lock.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Install deps for package.json without package-lock.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Install deps for package.json with yarn.lock", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("yarn.lock", []byte("{}"))

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "yarn", Params: []string{"install", "--frozen-lockfile"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Install deps for package.json with pnpm-lock.yaml", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))

		// Mock expects absolute path for locally installed pnpm
		absolutePnpmPath := "/tmp/node_modules/.bin/pnpm"
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			"pnpm --version":                fmt.Errorf("pnpm not installed globally"),
			absolutePnpmPath + " --version": fmt.Errorf("pnpm not installed locally"),
		}

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 5, len(utils.execRunner.Calls)) {
				// Set npm registry
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry", "-ws=false", "-iwr"}}, utils.execRunner.Calls[0])
				// Check global pnpm version command
				assert.Equal(t, mock.ExecCall{Exec: "pnpm", Params: []string{"--version"}}, utils.execRunner.Calls[1])
				// Check local pnpm version command
				assert.Equal(t, mock.ExecCall{Exec: absolutePnpmPath, Params: []string{"--version"}}, utils.execRunner.Calls[2])
				// Check pnpm install command
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "pnpm", "--prefix", "/tmp"}}, utils.execRunner.Calls[3])
				// Check pnpm install --frozen-lockfile command
				assert.Equal(t, mock.ExecCall{Exec: absolutePnpmPath, Params: []string{"install", "--frozen-lockfile"}}, utils.execRunner.Calls[4])
			}
		}
	})

	t.Run("Install all deps", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.InstallAllDependencies([]string{"package.json", filepath.Join("src", "package.json")})

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("check Execute script", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.executeScript("package.json", "ci-lint", []string{"--silent"}, []string{"--tag", "tag1"})

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent", "--", "--tag", "tag1"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("check Execute all scripts", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))

		options := ExecutorOptions{}
		runScripts := []string{"ci-lint", "ci-build"}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.RunScriptsInAllPackages(runScripts, nil, nil, false, nil, nil)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-build"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("check Execute all scripts with buildDescriptorList", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))                        // is filtered out
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }")) // should NOT be filtered out

		options := ExecutorOptions{}
		runScripts := []string{"ci-lint", "ci-build"}
		buildDescriptorList := []string{filepath.Join("src", "package.json"), "package.json"}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.RunScriptsInAllPackages(runScripts, nil, nil, false, nil, buildDescriptorList)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("check set npm registry", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))
		utils.execRunner = &mock.ExecMockRunner{StdoutReturn: map[string]string{"npm config get registry -ws=false -iwr": "undefined"}}
		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "https://example.org/npm"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.SetNpmRegistries()

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "get", "registry", "-ws=false", "-iwr"}}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"config", "set", "registry", exec.Options.DefaultNpmRegistry, "-ws=false", "-iwr"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Call run-scripts with virtual frame buffer", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"foo\": \"\" } }"))

		options := ExecutorOptions{}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
			pnpmSetup: pnpmSetupState{
				rootDir: "/", // Mock root directory
			},
		}
		err := exec.RunScriptsInAllPackages([]string{"foo"}, nil, nil, true, nil, nil)

		assert.Contains(t, utils.execRunner.Env, "DISPLAY=:99")
		assert.NoError(t, err)
		if assert.Len(t, utils.execRunner.Calls, 3) {
			xvfbCall := utils.execRunner.Calls[0]
			assert.Equal(t, "Xvfb", xvfbCall.Exec)
			assert.Equal(t, []string{"-ac", ":99", "-screen", "0", "1280x1024x16"}, xvfbCall.Params)
			assert.True(t, xvfbCall.Async)
			assert.True(t, xvfbCall.Execution.Killed)

			assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "foo"}}, utils.execRunner.Calls[2])
		}
	})

}
