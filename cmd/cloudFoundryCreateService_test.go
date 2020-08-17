package cmd

import (
	"io/ioutil"
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
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
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
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-t", "testTag, testTag2"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
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
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-b", "testBroker"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
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
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service", "testService", "testPlan", "testName", "-c", "testConfig.json"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})

	t.Run("Create service: failure, no config", func(t *testing.T) {
		defer cfMockCleanup(m)
		config := cloudFoundryCreateServiceOptions{}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		assert.EqualError(t, error, "Error while logging in: Failed to login to Cloud Foundry: Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password")
	})

	t.Run("Create service: variable substitution", func(t *testing.T) {
		defer cfMockCleanup(m)
		var manifestVariables = []string{"name1=Test1", "name2=Test2"}
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			ServiceManifest:   "testManifest",
			ManifestVariables: manifestVariables,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", "testManifest", "--var", "name1=Test1", "--var", "name2=Test2"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})

	t.Run("Create service: variable substitution with manifest file", func(t *testing.T) {
		defer cfMockCleanup(m)

		dir, err := ioutil.TempDir("", "test get result ATC run")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()
		bodyString := `name: test1
		name2: test2`
		body := []byte(bodyString)
		err = ioutil.WriteFile("file.test", body, 0644)
		err = ioutil.WriteFile("file2.test", body, 0644)

		var manifestVariablesFiles = []string{"file.test", "file2.test"}
		config := cloudFoundryCreateServiceOptions{
			CfAPIEndpoint:          "https://api.endpoint.com",
			CfOrg:                  "testOrg",
			CfSpace:                "testSpace",
			Username:               "testUser",
			Password:               "testPassword",
			ServiceManifest:        "testManifest",
			ManifestVariablesFiles: manifestVariablesFiles,
		}
		error := runCloudFoundryCreateService(&config, &telemetryData, cf)
		if assert.NoError(t, error) {
			assert.Equal(t, []mock.ExecCall{mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"create-service-push", "--no-push", "--service-manifest", "testManifest", "--vars-file", "file.test", "--vars-file", "file2.test"}},
				mock.ExecCall{Execution: (*mock.Execution)(nil), Async: false, Exec: "cf", Params: []string{"logout"}}},
				m.Calls)
		}
	})
}
