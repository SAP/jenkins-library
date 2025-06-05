//go:build unit

package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func cfMockCleanup(m *mock.ExecMockRunner) {
	m.ShouldFailOnCommand = map[string]error{}
	m.StdoutReturn = map[string]string{}
	m.Calls = []mock.ExecCall{}
}

func TestCloudFoundryCreateService(t *testing.T) {

	m := &mock.ExecMockRunner{}
	cf := cloudfoundry.CFUtils{Exec: m}

	var telemetryData telemetry.CustomData
	t.Run("Create service: no broker, no config, no tags", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:         "https://api.endpoint.com",
			CfOrg:                 "testOrg",
			CfSpace:               "testSpace",
			Username:              "testUser",
			Password:              "testPassword",
			CfService:             "testService",
			CfServiceInstanceName: "testName",
			CfServicePlan:         "testPlan",
			CfAsync:               false,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "--wait"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
	t.Run("Create service: only tags", func(t *testing.T) {
		defer cfMockCleanup(m)

		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:         "https://api.endpoint.com",
			CfOrg:                 "testOrg",
			CfSpace:               "testSpace",
			Username:              "testUser",
			Password:              "testPassword",
			CfService:             "testService",
			CfServiceInstanceName: "testName",
			CfServicePlan:         "testPlan",
			CfServiceTags:         "testTag, testTag2",
			CfAsync:               true,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-t", "testTag, testTag2"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
	t.Run("Create service: only broker", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:         "https://api.endpoint.com",
			CfOrg:                 "testOrg",
			CfSpace:               "testSpace",
			Username:              "testUser",
			Password:              "testPassword",
			CfService:             "testService",
			CfServiceInstanceName: "testName",
			CfServicePlan:         "testPlan",
			CfServiceBroker:       "testBroker",
			CfAsync:               true,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-b", "testBroker"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
	t.Run("Create service: only config", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:         "https://api.endpoint.com",
			CfOrg:                 "testOrg",
			CfSpace:               "testSpace",
			Username:              "testUser",
			Password:              "testPassword",
			CfService:             "testService",
			CfServiceInstanceName: "testName",
			CfServicePlan:         "testPlan",
			CfCreateServiceConfig: "testConfig.json",
			CfAsync:               true,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-c", "testConfig.json"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})

	t.Run("Create service: failure, no config", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := cloudFoundryCreateServiceOptions{}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		assert.EqualError(t, error, "Error while logging in: Failed to login to Cloud Foundry: Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password")
	})

	t.Run("Create service: variable substitution in-line", func(t *testing.T) {
		defer cfMockCleanup(m)

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
		---
		create-services:
		- name:   ((name))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   ((name2))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   "test3"
		  broker: "testBroker"
		  plan:   "testPlan"`

		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("manifestTest.yml", manifestFileStringBody, 0644)
		assert.NoError(t, err)

		var manifestVariables = []string{"name1=Test1", "name2=Test2"}

		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			ServiceManifest:   "manifestTest.yml",
			ManifestVariables: manifestVariables,
			CfAsync:           false, // should be ignored
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", "manifestTest.yml", "--var", "name1=Test1", "--var", "name2=Test2"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})

	t.Run("Create service: variable substitution with variable substitution manifest file", func(t *testing.T) {
		defer cfMockCleanup(m)

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()
		varsFileString := `name: test1
		name2: test2`

		manifestFileString := `
		---
		create-services:
		- name:   ((name))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   ((name2))
		  broker: "testBroker"
		  plan:   "testPlan"

		- name:   "test3"
		  broker: "testBroker"
		  plan:   "testPlan"`

		varsFileStringBody := []byte(varsFileString)
		manifestFileStringBody := []byte(manifestFileString)
		err := os.WriteFile("varsTest.yml", varsFileStringBody, 0644)
		assert.NoError(t, err)
		err = os.WriteFile("varsTest2.yml", varsFileStringBody, 0644)
		assert.NoError(t, err)
		err = os.WriteFile("manifestTest.yml", manifestFileStringBody, 0644)
		assert.NoError(t, err)

		var manifestVariablesFiles = []string{"varsTest.yml", "varsTest2.yml"}
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:          "https://api.endpoint.com",
			CfOrg:                  "testOrg",
			CfSpace:                "testSpace",
			Username:               "testUser",
			Password:               "testPassword",
			ServiceManifest:        "manifestTest.yml",
			ManifestVariablesFiles: manifestVariablesFiles,
			ManifestVariables:      []string{"a=b", "x=y"},
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", "manifestTest.yml", "--vars-file", "varsTest.yml", "--vars-file", "varsTest2.yml", "--var", "a=b", "--var", "x=y"}},
				{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
}
