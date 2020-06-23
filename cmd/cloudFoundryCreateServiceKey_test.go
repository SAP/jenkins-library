package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCloudFoundryCreateServiceKey(t *testing.T) {
	execRunner := mock.ExecMockRunner{}
	var telemetryData telemetry.CustomData
	t.Run("CF Login: success case", func(t *testing.T) {
		loginconfig := cloudFoundryDeleteServiceOptions{
			CfAPIEndpoint: "https://api.endpoint.com",
			CfOrg:         "testOrg",
			CfSpace:       "testSpace",
			Username:      "testUser",
			Password:      "testPassword",
		}
		error := cloudFoundryLogin(loginconfig, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[0].Exec)
			assert.Equal(t, []string{"login", "-a", "https://api.endpoint.com", "-o", "testOrg", "-s", "testSpace", "-u", "testUser", "-p", "testPassword"}, execRunner.Calls[0].Params)
		}
	})
	t.Run("CF Create Service Key: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			Username:          "testUser",
			Password:          "testPassword",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testKey",
		}
		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[1].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey"}, execRunner.Calls[1].Params)
		}
	})
	t.Run("CF Create Service Key with service Key config: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "testconfig.yml",
		}
		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[2].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", "testconfig.yml"}, execRunner.Calls[2].Params)
		}
	})
	t.Run("CF Create Service Key with service Key config: Success case", func(t *testing.T) {
		config := cloudFoundryCreateServiceKeyOptions{
			CfAPIEndpoint:      "https://api.endpoint.com",
			CfOrg:              "testOrg",
			CfSpace:            "testSpace",
			Username:           "testUser",
			Password:           "testPassword",
			CfServiceInstance:  "testInstance",
			CfServiceKeyName:   "testKey",
			CfServiceKeyConfig: "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}",
		}
		error := runCloudFoundryCreateServiceKey(&config, &telemetryData, &execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[3].Exec)
			assert.Equal(t, []string{"create-service-key", "testInstance", "testKey", "-c", "{\"scenario_id\":\"SAP_COM_0510\",\"type\":\"basic\"}"}, execRunner.Calls[3].Params)
		}
	})
	t.Run("CF Logout: Success case", func(t *testing.T) {
		error := cloudFoundryLogout(&execRunner)
		if error == nil {
			assert.Equal(t, "cf", execRunner.Calls[4].Exec)
			assert.Equal(t, "logout", execRunner.Calls[4].Params[0])
		}
	})
}
