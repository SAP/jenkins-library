//go:build unit
// +build unit

package whitesource

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteUAConfigurationFile(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:      "npm",
			ConfigFilePath: "ua.props",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile(config.ConfigFilePath, []byte("test = dummy"))

		path, err := config.RewriteUAConfigurationFile(utilsMock, "", false)
		assert.NoError(t, err)
		newUAConfig, err := utilsMock.FileRead(path)
		assert.NoError(t, err)
		assert.Contains(t, string(newUAConfig), "test = dummy")
		assert.Contains(t, string(newUAConfig), "failErrorLevel = ALL")
	})

	t.Run("accept non-existing file", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:      "npm",
			ConfigFilePath: "ua.props",
		}
		utilsMock := NewScanUtilsMock()

		path, err := config.RewriteUAConfigurationFile(utilsMock, "", false)
		assert.NoError(t, err)

		newUAConfig, err := utilsMock.FileRead(path)
		assert.NoError(t, err)
		assert.Contains(t, string(newUAConfig), "failErrorLevel = ALL")
	})

	t.Run("error - write file", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:      "npm",
			ConfigFilePath: "ua.props",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.FileWriteError = fmt.Errorf("failed to write file")

		_, err := config.RewriteUAConfigurationFile(utilsMock, "", false)
		assert.Contains(t, fmt.Sprint(err), "failed to write file")
	})
}

func TestUpdateConfig(t *testing.T) {
	t.Parallel()

	originalConfig := map[string]string{
		"dependent":            "dependentValue",
		"forced":               "forced_original",
		"not_forced":           "not_forced_original",
		"dont_omit_forced":     "dont_omit_forced_original",
		"dont_omit_not_forced": "dont_omit_not_forced_original",
		"append":               "original_value appended by",
		"append_empty":         "",
	}
	testConfig := ConfigOptions{
		{Name: "non_existing_forced", Value: "non_existing_forced_val", Force: true},
		{Name: "non_existing_not_forced", Value: "non_existing_not_forced_val", Force: false},
		{Name: "forced", Value: "forced_val", Force: true},
		{Name: "not_forced", Value: "not_forced_val", Force: false},
		{Name: "omit", Value: "omit_val", OmitIfPresent: "dependent"},
		{Name: "dont_omit", Value: "dont_omit_val", OmitIfPresent: "dependent_notExisting"},
		{Name: "dont_omit_forced", Value: "dont_omit_forced_val", OmitIfPresent: "dependent_notExisting", Force: true},
		{Name: "dont_omit_not_forced", Value: "dont_omit_not_forced_val", OmitIfPresent: "dependent_notExisting", Force: false},
		{Name: "append", Value: "appended_val", Append: true},
		{Name: "append_empty", Value: "appended_val", Append: true},
	}

	updatedConfig := testConfig.updateConfig(&originalConfig)

	assert.Equal(t, "dependentValue", updatedConfig["dependent"])
	assert.Equal(t, "non_existing_forced_val", updatedConfig["non_existing_forced"])
	assert.Equal(t, "non_existing_not_forced_val", updatedConfig["non_existing_not_forced"])
	assert.Equal(t, "forced_val", updatedConfig["forced"])
	assert.Equal(t, "not_forced_original", updatedConfig["not_forced"])
	assert.NotEqual(t, "omit_val", updatedConfig["omit"])
	assert.Equal(t, "dont_omit_val", updatedConfig["dont_omit"])
	assert.Equal(t, "dont_omit_forced_val", updatedConfig["dont_omit_forced"])
	assert.Equal(t, "dont_omit_not_forced_original", updatedConfig["dont_omit_not_forced"])
	assert.Equal(t, "original_value appended by appended_val", updatedConfig["append"])
	assert.Equal(t, "appended_val", updatedConfig["append_empty"])
}

func TestAddGeneralDefaults(t *testing.T) {
	t.Parallel()

	utilsMock := NewScanUtilsMock()

	t.Run("default", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			OrgToken:       "testOrgToken",
			ProductName:    "Test",
			ProductToken:   "testProductToken",
			ProductVersion: "testVersion",
			ProjectName:    "testProject",
			UserToken:      "testuserKey",
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock, "")
		assert.Equal(t, "checkPolicies", testConfig[0].Name)
		assert.Equal(t, true, testConfig[0].Value)
		assert.Equal(t, "forceCheckAllDependencies", testConfig[1].Name)
		assert.Equal(t, true, testConfig[1].Value)

		assert.Equal(t, "testOrgToken", testConfig[2].Value)
		assert.Equal(t, "Test", testConfig[3].Value)
		assert.Equal(t, "testVersion", testConfig[4].Value)
		assert.Equal(t, "testProject", testConfig[5].Value)
		assert.Equal(t, "testVersion", testConfig[6].Value)
		assert.Equal(t, "testProductToken", testConfig[7].Value)
		assert.Equal(t, "testuserKey", testConfig[8].Value)
	})

	t.Run("DIST product", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			OrgToken:       "testOrgToken",
			ProductName:    "DIST - Test",
			ProductToken:   "testProductToken",
			ProductVersion: "testVersion",
			ProjectName:    "testProject",
			UserToken:      "testuserKey",
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock, "anotherProject")
		assert.Equal(t, "checkPolicies", testConfig[0].Name)
		assert.Equal(t, false, testConfig[0].Value)
		assert.Equal(t, "forceCheckAllDependencies", testConfig[1].Name)
		assert.Equal(t, false, testConfig[1].Value)
		assert.Equal(t, "anotherProject", testConfig[5].Value)
	})

	t.Run("verbose", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			Verbose: true,
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock, "")
		assert.Equal(t, "log.level", testConfig[2].Name)
		assert.Equal(t, "debug", testConfig[2].Value)
		assert.Equal(t, "log.files.level", testConfig[3].Name)
		assert.Equal(t, "debug", testConfig[3].Value)
	})

	t.Run("includes and excludes", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			Excludes: []string{"**/excludes1", "**/excludes2"},
			Includes: []string{"**/includes1", "**/includes2"},
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock, "")
		assert.Equal(t, "excludes", testConfig[2].Name)
		assert.Equal(t, "**/excludes1 **/excludes2", testConfig[2].Value)
		assert.Equal(t, true, testConfig[2].Force)
		assert.Equal(t, "includes", testConfig[3].Name)
		assert.Equal(t, "**/includes1 **/includes2", testConfig[3].Value)
		assert.Equal(t, true, testConfig[3].Force)
	})
}

func TestAddBuildToolDefaults(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		var testConfig ConfigOptions
		whitesourceConfig := ScanOptions{
			BuildTool: "dub",
		}
		err := testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, ConfigOptions{{Name: "ignoreSourceFiles", Value: true, Force: true}, {Name: "includes", Value: "**/*.d **/*.di"}}, testConfig)
	})

	t.Run("success case", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		var testConfig ConfigOptions
		whitesourceConfig := ScanOptions{
			BuildTool: "dub2",
		}
		err := testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, ConfigOptions{{Name: "fileSystemScan", Value: false, Force: true}, {Name: "includes", Value: "**/*.d **/*.di"}}, testConfig)
	})

	t.Run("error case", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		var testConfig ConfigOptions
		whitesourceConfig := ScanOptions{
			BuildTool: "notHardened",
		}
		err := testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.EqualError(t, err, "configuration not hardened")
	})

	t.Run("maven - m2 path", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool: "maven",
			M2Path:    "test/.m2",
		}
		testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.Contains(t, testConfig, ConfigOption{Name: "maven.m2RepositoryPath", Value: "test/.m2", Force: true})
		assert.NotContains(t, testConfig, ConfigOption{Name: "maven.additionalArguments", Value: "", Force: true})
	})

	t.Run("maven - settings", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool:                  "maven",
			ProjectSettingsFile:        "project-settings.xml",
			GlobalSettingsFile:         "global-settings.xml",
			BuildDescriptorExcludeList: []string{"unit-tests/pom.xml"},
		}
		utilsMock.AddFile("unit-tests/pom.xml", []byte("dummy"))
		testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		dir, _ := os.Getwd()
		globalSettingsPath := filepath.Join(dir, "global-settings.xml")
		projectSettingsPath := filepath.Join(dir, "project-settings.xml")
		assert.Contains(t, testConfig, ConfigOption{Name: "maven.additionalArguments", Value: "--global-settings " + globalSettingsPath + " --settings " + projectSettingsPath + " --projects !unit-tests", Append: true})
	})

	t.Run("Docker - default", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool: "docker",
		}
		utilsMock.AddFile("Dockerfile", []byte("dummy"))
		testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.Contains(t, testConfig, ConfigOption{Name: "docker.dockerfilePath", Value: "Dockerfile", Force: false})
	})

	t.Run("Docker - no builddescriptor found", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool: "docker",
		}
		testConfig.addBuildToolDefaults(&whitesourceConfig, utilsMock)
		assert.NotContains(t, testConfig, ConfigOption{Name: "docker.dockerfilePath", Value: "Dockerfile", Force: false})
	})
}

func TestMvnProjectExcludes(t *testing.T) {
	utilsMock := NewScanUtilsMock()
	utilsMock.AddFile("unit-tests/package.json", []byte("dummy"))
	utilsMock.AddFile("unit-tests/pom.xml", []byte("dummy"))
	utilsMock.AddFile("integration-tests/pom.xml", []byte("dummy"))
	tt := []struct {
		buildDescriptorExcludeList []string
		expected                   []string
	}{
		{buildDescriptorExcludeList: []string{}, expected: []string{}},
		{buildDescriptorExcludeList: []string{"unit-tests/package.json", "integration-tests/package.json"}, expected: []string{}},
		{buildDescriptorExcludeList: []string{"unit-tests/pom.xml"}, expected: []string{"--projects", "!unit-tests"}},
		{buildDescriptorExcludeList: []string{"unit-tests/pom.xml", "integration-tests/pom.xml"}, expected: []string{"--projects", "!unit-tests,!integration-tests"}},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, mvnProjectExcludes(test.buildDescriptorExcludeList, utilsMock), test.buildDescriptorExcludeList)
	}
}
