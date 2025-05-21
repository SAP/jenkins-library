//go:build unit
// +build unit

package npm

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type npmMockUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

func (u *npmMockUtilsBundle) GetExecRunner() ExecRunner {
	return u.execRunner
}

func (u *npmMockUtilsBundle) LookPath(a string) (string, error) {
	return a, nil
}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestNpm(t *testing.T) {
	t.Run("find package.json files with one package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))

		options := ExecutorOptions{}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
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

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
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

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
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

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
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

		options := ExecutorOptions{DefaultNpmRegistry: "foo.bar", Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				fmt.Println(utils.execRunner.Calls)
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Install deps for package.json without package-lock.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := ExecutorOptions{DefaultNpmRegistry: "foo.bar", Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.install("package.json")

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
		utils.execRunner.Lookups = map[string]bool{"yarn": true} // assume yarn is pre-installed

		options := ExecutorOptions{DefaultNpmRegistry: "foo.bar", Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.install("package.json")

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: tool.GetBinaryPath(), Params: []string{"install", "--frozen-lockfile"}}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Install all deps", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))

		options := ExecutorOptions{DefaultNpmRegistry: "foo.bar", Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.InstallAllDependencies([]string{"package.json", filepath.Join("src", "package.json")})

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"ci"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("check if yarn.lock exist", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("yarn.lock", []byte("{}"))
		utils.AddFile("package-lock.json", []byte("{}"))

		tool, err := DetectTool(&utils, "auto")
		if assert.NoError(t, err) {
			assert.Equal(t, "yarn", tool.Name)
		}
	})

	t.Run("for usage of npm", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))

		tool, err := DetectTool(&utils, "auto")
		if assert.NoError(t, err) {
			assert.Equal(t, "npm", tool.Name)
		}
	})

	t.Run("check if pnpm-lock.yaml exist", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))
		utils.AddFile("package-lock.json", []byte("{}"))

		tool, err := DetectTool(&utils, "auto")
		if assert.NoError(t, err) {
			assert.Equal(t, "pnpm", tool.Name)
		}
	})

	t.Run("check that yarn.lock and package-lock and pnpm-lock do not exist", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		tool, err := DetectTool(&utils, "auto")
		if assert.NoError(t, err) {
			assert.Equal(t, "npm", tool.Name)
		}
	})

	t.Run("use yarn when requested even without lock file", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.execRunner.Lookups = map[string]bool{"yarn": false}

		tool, err := DetectTool(&utils, "yarn")
		if assert.NoError(t, err) {
			assert.Equal(t, "yarn", tool.Name)
			if assert.Equal(t, 1, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "yarn", "--prefix", "./tmp"}}, utils.execRunner.Calls[0])
			}
		}
	})

	t.Run("use pnpm when requested even without lock file", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.execRunner.Lookups = map[string]bool{"pnpm": false}

		tool, err := DetectTool(&utils, "pnpm")
		if assert.NoError(t, err) {
			assert.Equal(t, "pnpm", tool.Name)
			if assert.Equal(t, 1, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "pnpm", "--prefix", "./tmp"}}, utils.execRunner.Calls[0])
			}
		}
	})

	t.Run("test auto-installer", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))
		utils.execRunner.Lookups = map[string]bool{"pnpm": false}

		tool, err := DetectTool(&utils, "auto")
		if assert.NoError(t, err) {
			assert.Equal(t, "pnpm", tool.Name)
			if assert.Equal(t, 1, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "pnpm", "--prefix", "./tmp"}}, utils.execRunner.Calls[0])
			}
		}
	})

	t.Run("check Execute script", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))

		options := ExecutorOptions{Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.executeScript("package.json", "ci-lint", []string{"--silent"}, []string{"--tag", "tag1"})

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent", "--", "--tag", "tag1"}}, utils.execRunner.Calls[1])
			}
		}
	})

	t.Run("check Execute script with pnpm", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))
		utils.execRunner.Lookups = map[string]bool{"pnpm": true} // assume pnpm is pre-installed

		options := ExecutorOptions{Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.executeScript("package.json", "ci-lint", []string{"--silent"}, []string{"--tag", "tag1"})

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: tool.GetBinaryPath(), Params: []string{"run", "ci-lint", "--silent", "--", "--tag", "tag1"}}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("check Execute all scripts", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-build\": \"exit 0\" } }"))

		options := ExecutorOptions{Tool: "auto"}
		runScripts := []string{"ci-lint", "ci-build"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.RunScriptsInAllPackages(runScripts, nil, nil, false, nil, nil)

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

		options := ExecutorOptions{Tool: "auto"}
		runScripts := []string{"ci-lint", "ci-build"}
		buildDescriptorList := []string{filepath.Join("src", "package.json"), "package.json"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.RunScriptsInAllPackages(runScripts, nil, nil, false, nil, buildDescriptorList)

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

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.SetNpmRegistries()

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

		options := ExecutorOptions{Tool: "auto"}

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.RunScriptsInAllPackages([]string{"foo"}, nil, nil, true, nil, nil)

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

	t.Run("Create BOM with cyclonedx-npm", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.CreateBOM([]string{"package.json", filepath.Join("src", "package.json")})
		cycloneDxNpmInstallParams := []string{"install", "--no-save", "@cyclonedx/cyclonedx-npm@1.11.0", "--prefix", "./tmp"}
		cycloneDxNpmRunParams := []string{
			"--output-format",
			"XML",
			"--spec-version",
			cycloneDxSchemaVersion,
			"--omit",
			"dev",
			"--output-file",
		}

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: cycloneDxNpmInstallParams}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "./tmp/node_modules/.bin/cyclonedx-npm", Params: append(cycloneDxNpmRunParams, "bom-npm.xml", "package.json")}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "./tmp/node_modules/.bin/cyclonedx-npm", Params: append(cycloneDxNpmRunParams, filepath.Join("src", "bom-npm.xml"), filepath.Join("src", "package.json"))}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Create BOM with fallback cyclonedx/bom", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))
		utils.execRunner.ShouldFailOnCommand = map[string]error{"npm install --no-save @cyclonedx/cyclonedx-npm@1.11.0 --prefix ./tmp": fmt.Errorf("failed to install CycloneDX BOM")}

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		tool, err := DetectTool(&utils, "auto")
		assert.NoError(t, err)
		exec := &Execute{
			Utils:   &utils,
			Options: options,
			Tool:    tool,
		}
		err = exec.CreateBOM([]string{"package.json", filepath.Join("src", "package.json")})
		cycloneDxNpmInstallParams := []string{"install", "--no-save", "@cyclonedx/cyclonedx-npm@1.11.0", "--prefix", "./tmp"}

		cycloneDxBomInstallParams := []string{"install", cycloneDxBomPackageVersion, "--no-save"}
		cycloneDxBomRunParams := []string{
			"cyclonedx-bom",
			"--output",
		}

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: cycloneDxNpmInstallParams}, utils.execRunner.Calls[0])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: cycloneDxBomInstallParams}, utils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: append(cycloneDxBomRunParams, "bom-npm.xml", ".")}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: append(cycloneDxBomRunParams, filepath.Join("src", "bom-npm.xml"), filepath.Join("src"))}, utils.execRunner.Calls[3])
			}
		}
	})
}
