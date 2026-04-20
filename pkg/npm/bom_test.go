//go:build unit
// +build unit

package npm

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestBom(t *testing.T) {
	t.Run("Create BOM with cyclonedx-npm", func(t *testing.T) {
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
		}
		err := exec.CreateBOM([]string{"package.json", filepath.Join("src", "package.json")})
		cycloneDxNpmInstallParams := []string{"install", "--no-save", "@cyclonedx/cyclonedx-npm@2.1.0", "--prefix", "./tmp"}
		cycloneDxNpmRunParams := []string{
			"--output-format",
			"XML",
			"--spec-version",
			CycloneDxSchemaVersion,
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

	t.Run("Create BOM fails if cyclonedx-npm install fails", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile(filepath.Join("src", "package-lock.json"), []byte("{}"))
		utils.execRunner.ShouldFailOnCommand = map[string]error{"npm install --no-save @cyclonedx/cyclonedx-npm@2.1.0 --prefix ./tmp": fmt.Errorf("failed to install CycloneDX BOM")}

		options := ExecutorOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		exec := &Execute{
			Utils:   &utils,
			Options: options,
		}
		err := exec.CreateBOM([]string{"package.json", filepath.Join("src", "package.json")})

		assert.Contains(t, err.Error(), "failed to install CycloneDX BOM")
		// Only the failing npm install should have been called; no @cyclonedx/bom or npx invocation
		if assert.Equal(t, 1, len(utils.execRunner.Calls)) {
			assert.NotContains(t, utils.execRunner.Calls[0].Exec, "@cyclonedx/bom")
			assert.NotContains(t, utils.execRunner.Calls[0].Exec, "npx")
		}
	})

	t.Run("Create BOM fails if cyclonedx-npm run step fails", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("package-lock.json", []byte("{}"))

		// npm install succeeds, but the cyclonedx-npm executable itself fails
		runCmd := strings.Join([]string{
			"./tmp/node_modules/.bin/cyclonedx-npm",
			"--output-format", "XML",
			"--spec-version", CycloneDxSchemaVersion,
			"--omit", "dev",
			"--output-file", "bom-npm.xml",
			"package.json",
		}, " ")
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			runCmd: fmt.Errorf("cyclonedx-npm execution failed"),
		}

		options := ExecutorOptions{DefaultNpmRegistry: "foo.bar"}
		exec := &Execute{
			Utils:   &utils,
			Options: options,
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "cyclonedx-npm execution failed")
	})

	t.Run("Create BOM with cdxgen for pnpm project", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"exit 0\" } }"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))

		options := ExecutorOptions{
			DefaultNpmRegistry: "foo.bar",
		}

		exec := &Execute{
			Utils:   &utils,
			Options: options,
		}

		// Execute CreateBOM
		err := exec.CreateBOM([]string{"package.json"})

		if assert.NoError(t, err) {
			url := cycloneDxCliUrl[struct{ os, arch string }{runtime.GOOS, runtime.GOARCH}]
			cyclonedxExec := filepath.Join(".pipeline", filepath.Base(url))
			// Verify cyclonedx-cli download was requested
			assert.True(t, utils.FilesMock.HasFile(cyclonedxExec))

			// Verify command execution sequence
			assert.Equal(t, 4, len(utils.execRunner.Calls))

			// Check pnpm version check call
			assert.Equal(t, mock.ExecCall{
				Exec:   "pnpm",
				Params: []string{"--version"},
			}, utils.execRunner.Calls[0])

			// Check cdxgen install
			assert.Equal(t, mock.ExecCall{
				Exec:   "npm",
				Params: []string{"install", cdxgenPackageVersion, "--prefix", tmpInstallFolder},
			}, utils.execRunner.Calls[1])

			// Check cdxgen execution generating JSON
			cdxgenPath := tmpInstallFolder + "/node_modules/.bin/cdxgen"
			assert.Equal(t, mock.ExecCall{
				Exec: cdxgenPath,
				Params: []string{
					"-r",
					"-o", "bom-npm.json",
					"--spec-version", CycloneDxSchemaVersion,
				},
			}, utils.execRunner.Calls[2])

			// Check cyclonedx-cli conversion from JSON to XML
			// The output version for cyclonedx-cli is expected to be in the format "vX_Y_Z". Ex: 1.4 => v1_4
			outputVersion := fmt.Sprintf("v%s", strings.ReplaceAll(CycloneDxSchemaVersion, ".", "_"))
			assert.Equal(t, mock.ExecCall{
				Exec:   cyclonedxExec,
				Params: []string{"convert", "--input-file", "bom-npm.json", "--output-format", "xml", "--output-file", "bom-npm.xml", "--output-version", outputVersion},
			}, utils.execRunner.Calls[3])
		}
	})

	t.Run("Create BOM fails if cdxgen install fails", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			"npm install @cyclonedx/cdxgen@12.1.3 --prefix ./tmp": fmt.Errorf("failed to install cdxgen"),
		}

		exec := &Execute{
			Utils:   &utils,
			Options: ExecutorOptions{},
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "failed to install cdxgen")
	})

	t.Run("Create BOM fails if cdxgen execution fails", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))

		// cdxgen install succeeds, but cdxgen executable itself fails when generating the BOM
		cdxgenCmd := strings.Join([]string{
			"./tmp/node_modules/.bin/cdxgen",
			"-r",
			"-o", "bom-npm.json",
			"--spec-version", CycloneDxSchemaVersion,
		}, " ")
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			cdxgenCmd: fmt.Errorf("cdxgen execution failed"),
		}

		exec := &Execute{
			Utils:   &utils,
			Options: ExecutorOptions{},
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "cdxgen execution failed")
	})

	t.Run("Create BOM fails if cyclonedx-cli conversion fails after the cdxgen generation", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))

		// cdxgen download, install and execution succeed, but the cyclonedx-cli convert step fails
		url := cycloneDxCliUrl[struct{ os, arch string }{runtime.GOOS, runtime.GOARCH}]
		cyclonedxExec := filepath.Join(".pipeline", filepath.Base(url))
		outputVersion := fmt.Sprintf("v%s", strings.ReplaceAll(CycloneDxSchemaVersion, ".", "_"))
		convertCmd := strings.Join([]string{
			cyclonedxExec,
			"convert",
			"--input-file", "bom-npm.json",
			"--output-format", "xml",
			"--output-file", "bom-npm.xml",
			"--output-version", outputVersion,
		}, " ")
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			convertCmd: fmt.Errorf("cyclonedx-cli conversion failed"),
		}

		exec := &Execute{
			Utils:   &utils,
			Options: ExecutorOptions{},
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "cyclonedx-cli conversion failed")

		// Verify cdxgen ran successfully before the conversion failure
		cdxgenPath := tmpInstallFolder + "/node_modules/.bin/cdxgen"
		if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
			assert.Equal(t, mock.ExecCall{
				Exec:   cdxgenPath,
				Params: []string{"-r", "-o", "bom-npm.json", "--spec-version", CycloneDxSchemaVersion},
			}, utils.execRunner.Calls[2])
			assert.Equal(t, mock.ExecCall{
				Exec:   cyclonedxExec,
				Params: []string{"convert", "--input-file", "bom-npm.json", "--output-format", "xml", "--output-file", "bom-npm.xml", "--output-version", outputVersion},
			}, utils.execRunner.Calls[3])
		}
	})

	t.Run("Create BOM fails if cyclonedx-cli download fails with network error", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))

		// Configure HttpClientMock to simulate network download failure
		utils.downloadClient.ReturnDownloadError = fmt.Errorf("network error: connection refused")

		exec := &Execute{
			Utils:   &utils,
			Options: ExecutorOptions{},
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "network error: connection refused")
	})
}
