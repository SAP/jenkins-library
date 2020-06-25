package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateService(t *testing.T) {
	execRunner := mock.ExecMockRunner{}
	var telemetryData telemetry.CustomData
	t.Run("Create service: no broker, no config, no tags", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"create-service", "testService", "testPlan", "testName"}, execRunner.Calls[0].Params)
		}
	})
	t.Run("Create service: only tags", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"create-service", "testService", "testPlan", "testName", "-t", "testTag, testTag2"}, execRunner.Calls[1].Params)
		}
	})
	t.Run("Create service: only broker", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, []string{"create-service", "testService", "testPlan", "testName", "-b", "testBroker"}, execRunner.Calls[2].Params)
		}
	})
	t.Run("Create service: only config", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"create-service", "testService", "testPlan", "testName", "-c", "testConfig.json"}, execRunner.Calls[3].Params)
		}
	})

	t.Run("Create service: failure, no config", func(t *testing.T) {
		config := cloudFoundryCreateServiceOptions{}
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[4].Exec)
			assert.Equal(t, []string{"create-service", "", "", ""}, execRunner.Calls[4].Params)
		}
	})
	t.Run("Create service: variable substitution", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[5].Exec)
			assert.Equal(t, []string{"create-service-push", "--no-push", "--service-manifest", "testManifest", "--var", "name1=Test1", "--var", "name2=Test2"}, execRunner.Calls[5].Params)
		}
	})
	t.Run("Create service: variable substitution with manifest file", func(t *testing.T) {
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
		error := runCloudFoundryCreateService(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[6].Exec)
			assert.Equal(t, []string{"create-service-push", "--no-push", "--service-manifest", "testManifest", "--vars-file", "file.test", "--vars-file", "file2.test"}, execRunner.Calls[6].Params)
		}
	})
}
