package whitesource

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteUAConfigurationFile(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)

		uaFile := filepath.Join(dir, "ua.props")
		ioutil.WriteFile(uaFile, []byte{}, 0666)

		config := ScanOptions{
			BuildTool:      "npm",
			ConfigFilePath: uaFile,
		}
		path, err := config.RewriteUAConfigurationFile()
		assert.NoError(t, err)

		newUAConfig, err := ioutil.ReadFile(path)
		assert.NoError(t, err)
		assert.Contains(t, string(newUAConfig), "failErrorLevel = ALL")
	})

	t.Run("accept non-existing file", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)

		uaFile := filepath.Join(dir, "ua_na.props")

		config := ScanOptions{
			BuildTool:      "npm",
			ConfigFilePath: uaFile,
		}
		path, err := config.RewriteUAConfigurationFile()
		assert.NoError(t, err)

		newUAConfig, err := ioutil.ReadFile(path)
		assert.NoError(t, err)
		assert.Contains(t, string(newUAConfig), "failErrorLevel = ALL")
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
		testConfig.addGeneralDefaults(&whitesourceConfig)
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

	t.Run("verbose", func(t *testing.T) {
		testConfig := ConfigOptions{}
		whitesourceConfig := ScanOptions{
			Verbose: true,
		}
		testConfig.addGeneralDefaults(&whitesourceConfig)
		assert.Equal(t, "log.level", testConfig[2].Name)
		assert.Equal(t, "debug", testConfig[2].Value)
		assert.Equal(t, "log.files.level", testConfig[3].Name)
		assert.Equal(t, "debug", testConfig[3].Value)
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
		testConfig.addGeneralDefaults(&whitesourceConfig)
		assert.Equal(t, "checkPolicies", testConfig[0].Name)
		assert.Equal(t, false, testConfig[0].Value)
		assert.Equal(t, "forceCheckAllDependencies", testConfig[1].Name)
		assert.Equal(t, false, testConfig[1].Value)
	})
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
