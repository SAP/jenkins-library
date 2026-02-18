//go:build unit

package npm

import (
	"fmt"
	"path/filepath"
	"runtime"
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

		exec := &Execute{
			Utils:   &utils,
			Options: options,
		}
		err := exec.CreateBOM([]string{"package.json", filepath.Join("src", "package.json")})
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
					"--spec-version", cycloneDxSchemaVersion,
				},
			}, utils.execRunner.Calls[2])

			// Check cyclonedx-cli conversion from JSON to XML
			assert.Equal(t, mock.ExecCall{
				Exec:   cyclonedxExec,
				Params: []string{"convert", "--input-file", "bom-npm.json", "--output-format", "xml", "--output-file", "bom-npm.xml"},
			}, utils.execRunner.Calls[3])
		}
	})

	t.Run("Create BOM fails if cdxgen install fails", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{}"))
		utils.AddFile("pnpm-lock.yaml", []byte("{}"))
		utils.execRunner.ShouldFailOnCommand = map[string]error{
			"npm install @cyclonedx/cdxgen@^11.3.2 --prefix ./tmp": fmt.Errorf("failed to install cdxgen"),
		}

		exec := &Execute{
			Utils:   &utils,
			Options: ExecutorOptions{},
		}
		err := exec.CreateBOM([]string{"package.json"})

		assert.Contains(t, err.Error(), "failed to install cdxgen")
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
