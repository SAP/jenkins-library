package whitesource

import (
	"fmt"
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

		path, err := config.RewriteUAConfigurationFile(utilsMock)
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

		path, err := config.RewriteUAConfigurationFile(utilsMock)
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

		_, err := config.RewriteUAConfigurationFile(utilsMock)
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
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
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
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.Equal(t, "checkPolicies", testConfig[0].Name)
		assert.Equal(t, false, testConfig[0].Value)
		assert.Equal(t, "forceCheckAllDependencies", testConfig[1].Name)
		assert.Equal(t, false, testConfig[1].Value)
	})

	t.Run("verbose", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			Verbose: true,
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
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
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.Equal(t, "excludes", testConfig[2].Name)
		assert.Equal(t, "**/excludes1 **/excludes2", testConfig[2].Value)
		assert.Equal(t, true, testConfig[2].Force)
		assert.Equal(t, "includes", testConfig[3].Name)
		assert.Equal(t, "**/includes1 **/includes2", testConfig[3].Value)
		assert.Equal(t, true, testConfig[3].Force)
	})

	t.Run("maven - m2 path", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool: "maven",
			M2Path:    "test/.m2",
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.Equal(t, "maven.m2RepositoryPath", testConfig[2].Name)
		assert.Equal(t, "test/.m2", testConfig[2].Value)
		assert.Equal(t, true, testConfig[2].Force)
		assert.NotEqual(t, "maven.additionalArguments", testConfig[3].Name)
	})

	t.Run("maven - settings", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool:                  "maven",
			ProjectSettingsFile:        "project-settings.xml",
			GlobalSettingsFile:         "global-settings.xml",
			BuildDescriptorExcludeList: []string{"unit-tests/pom.xml"},
		}
		utilsMock.AddFile("unit-tests/pom.xml", []byte("dummy"))
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.Equal(t, "maven.additionalArguments", testConfig[2].Name)
		assert.Equal(t, "--global-settings global-settings.xml --settings project-settings.xml --projects !unit-tests", testConfig[2].Value)
		assert.Equal(t, true, testConfig[2].Force)
	})

	t.Run("Docker - default", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool: "docker",
		}
		utilsMock.AddFile("Dockerfile", []byte("dummy"))
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		// Name: "docker.dockerfilePath", Value: dockerFile, Force: false
		assert.Equal(t, "docker.dockerfilePath", testConfig[2].Name)
		assert.Equal(t, "Dockerfile", testConfig[2].Value)
		assert.Equal(t, false, testConfig[2].Force)
	})

	t.Run("Docker - custom", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool:           "docker",
			BuildDescriptorFile: "Dockerfile_custom",
		}
		utilsMock.AddFile("Dockerfile_custom", []byte("dummy"))
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.Equal(t, "docker.dockerfilePath", testConfig[2].Name)
		assert.Equal(t, "Dockerfile_custom", testConfig[2].Value)
		assert.Equal(t, false, testConfig[2].Force)
	})

	t.Run("Docker - no builddescriptor found", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			BuildTool:           "docker",
			BuildDescriptorFile: "Dockerfile_nonExisting",
		}
		testConfig.addGeneralDefaults(&whitesourceConfig, utilsMock)
		assert.NotEqual(t, "docker.dockerfilePath", testConfig[2].Name)
		assert.NotEqual(t, "Dockerfile_custom", testConfig[2].Value)
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

func TestAddBuildToolDefaults(t *testing.T) {
	t.Parallel()
	t.Run("success case", func(t *testing.T) {
		var testConfig ConfigOptions
		err := testConfig.addBuildToolDefaults("dub")
		assert.NoError(t, err)
		assert.Equal(t, ConfigOptions{{Name: "includes", Value: "**/*.d **/*.di"}}, testConfig)
	})

	t.Run("error case", func(t *testing.T) {
		var testConfig ConfigOptions
		err := testConfig.addBuildToolDefaults("not_available")
		assert.EqualError(t, err, "configuration not hardened")
	})
}
