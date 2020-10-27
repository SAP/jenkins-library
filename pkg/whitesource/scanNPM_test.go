package whitesource

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestExecuteScanNPM(t *testing.T) {
	config := ScanOptions{
		ScanType:    "npm",
		OrgToken:    "org-token",
		UserToken:   "user-token",
		ProductName: "mock-product",
		ProjectName: "mock-project",
	}

	t.Parallel()

	t.Run("happy path NPM", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteNpmScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		expectedCalls := []mock.ExecCall{
			{
				Exec: "npm",
				Params: []string{
					"ls",
				},
			},
			{
				Exec: "npx",
				Params: []string{
					"whitesource",
					"run",
				},
			},
		}
		assert.Equal(t, expectedCalls, utilsMock.Calls)
		assert.True(t, utilsMock.HasWrittenFile(whiteSourceConfig))
		assert.True(t, utilsMock.HasRemovedFile(whiteSourceConfig))
	})
	t.Run("no NPM modules", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteNpmScan(&config, utilsMock)
		// assert
		assert.EqualError(t, err, "found no NPM modules to scan. Configured excludes: []")
		assert.Len(t, utilsMock.Calls, 0)
		assert.False(t, utilsMock.HasWrittenFile(whiteSourceConfig))
	})
	t.Run("package.json needs name", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("package.json", []byte(`{"key":"value"}`))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteNpmScan(&config, utilsMock)
		// assert
		assert.EqualError(t, err, "failed to scan NPM module 'package.json': the file 'package.json' must configure a name")
	})
	t.Run("npm ls fails", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		utilsMock.AddFile(filepath.Join("app", "package.json"), []byte(`{"name":"my-app-module-name"}`))
		utilsMock.AddFile("package-lock.json", []byte("dummy"))

		utilsMock.ShouldFailOnCommand = make(map[string]error)
		utilsMock.ShouldFailOnCommand["npm ls"] = fmt.Errorf("mock failure")
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteNpmScan(&config, utilsMock)
		// assert
		assert.NoError(t, err)
		expectedNpmInstalls := []NpmInstall{
			{currentDir: "app", packageJSON: []string{"package.json"}},
			{currentDir: "", packageJSON: []string{"package.json"}},
		}
		assert.Equal(t, expectedNpmInstalls, utilsMock.NpmInstalledModules)
		assert.True(t, utilsMock.HasRemovedFile("package-lock.json"))
	})
}

func TestWriteWhitesourceConfigJSON(t *testing.T) {
	config := &ScanOptions{
		OrgToken:     "org-token",
		UserToken:    "user-token",
		ProductName:  "mock-product",
		ProjectName:  "mock-project",
		ProductToken: "mock-product-token",
	}

	expected := make(map[string]interface{})
	expected["apiKey"] = "org-token"
	expected["userKey"] = "user-token"
	expected["checkPolicies"] = true
	expected["productName"] = "mock-product"
	expected["projectName"] = "mock-project"
	expected["productToken"] = "mock-product-token"
	expected["productVer"] = "product-version"
	expected["devDep"] = true
	expected["ignoreNpmLsErrors"] = true

	t.Parallel()

	t.Run("write config from scratch", func(t *testing.T) {
		// init
		utils := NewScanUtilsMock()
		scan := newTestScan(config)
		// test
		err := scan.writeWhitesourceConfigJSON(config, utils, true, true)
		// assert
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(whiteSourceConfig)) {
			contents, _ := utils.FileRead(whiteSourceConfig)
			actual := make(map[string]interface{})
			_ = json.Unmarshal(contents, &actual)
			assert.Equal(t, expected, actual)
		}
	})

	t.Run("extend and merge config", func(t *testing.T) {
		// init
		initial := make(map[string]interface{})
		initial["checkPolicies"] = false
		initial["productName"] = "mock-product"
		initial["productVer"] = "41"
		initial["unknown"] = "preserved"
		encoded, _ := json.Marshal(initial)

		utils := NewScanUtilsMock()
		utils.AddFile(whiteSourceConfig, encoded)

		scan := newTestScan(config)

		// test
		err := scan.writeWhitesourceConfigJSON(config, utils, true, true)
		// assert
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(whiteSourceConfig)) {
			contents, _ := utils.FileRead(whiteSourceConfig)
			actual := make(map[string]interface{})
			_ = json.Unmarshal(contents, &actual)

			mergedExpected := expected
			mergedExpected["unknown"] = "preserved"

			assert.Equal(t, mergedExpected, actual)
		}
	})

	t.Run("extend and merge config, omit productToken", func(t *testing.T) {
		// init
		initial := make(map[string]interface{})
		initial["checkPolicies"] = false
		initial["productName"] = "mock-product"
		initial["productVer"] = "41"
		initial["unknown"] = "preserved"
		initial["projectToken"] = "mock-project-token"
		encoded, _ := json.Marshal(initial)

		utils := NewScanUtilsMock()
		utils.AddFile(whiteSourceConfig, encoded)

		scan := newTestScan(config)

		// test
		err := scan.writeWhitesourceConfigJSON(config, utils, true, true)
		// assert
		if assert.NoError(t, err) && assert.True(t, utils.HasWrittenFile(whiteSourceConfig)) {
			contents, _ := utils.FileRead(whiteSourceConfig)
			actual := make(map[string]interface{})
			_ = json.Unmarshal(contents, &actual)

			mergedExpected := expected
			mergedExpected["unknown"] = "preserved"
			mergedExpected["projectToken"] = "mock-project-token"
			delete(mergedExpected, "productToken")

			assert.Equal(t, mergedExpected, actual)
		}
	})
}
